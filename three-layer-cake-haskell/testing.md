# Testing the Business Layer

Companion to [SKILL.md](SKILL.md). Read this when the user asks how to test, or wants to verify the testability claim of the pattern.

The payoff of the three-layer cake is that business logic is **testable without IO**. The business functions are polymorphic in `m`; you can run them under any monad that satisfies the constraints.

---

## 1. Why this works

Recall the business function signature:

```haskell
createTodo
  :: (MonadTodoDb m, MonadLogger m, MonadError AppError m)
  => Text -> m Todo
```

`m` is a variable. In production it's `AppM` (which uses `IO`). In tests it can be any monad that has instances for `MonadTodoDb`, `MonadLogger`, and `MonadError AppError`. The most useful choice is a stack of `State` (for fake DB) and `Except` (for the error channel).

No `IO`, no `IORef`, no real database, no `tryIO`. The exact same business code runs.

---

## 2. The test monad

```haskell
{-# LANGUAGE GeneralizedNewtypeDeriving #-}
{-# LANGUAGE DerivingStrategies #-}

module Test.AppM where

import Control.Monad.Except
import Control.Monad.State
import Data.List (find)
import Data.Text (Text)

-- A minimal fake "world" — just the todos table.
data TestState = TestState
  { tsTodos    :: [Todo]
  , tsNextId   :: Int
  , tsLogs     :: [Text]  -- captured log lines for assertions
  } deriving Show

emptyState :: TestState
emptyState = TestState { tsTodos = [], tsNextId = 1, tsLogs = [] }

newtype TestM a = TestM
  { unTestM :: StateT TestState (Except AppError) a }
  deriving newtype
    ( Functor, Applicative, Monad
    , MonadState TestState
    , MonadError AppError
    )

runTestM :: TestState -> TestM a -> Either AppError (a, TestState)
runTestM s0 = runExcept . flip runStateT s0 . unTestM

-- Convenience: start from emptyState.
runTest :: TestM a -> Either AppError (a, TestState)
runTest = runTestM emptyState
```

That's the whole infrastructure. `State` carries the fake world. `Except` carries domain errors. No `IO` anywhere in the stack.

---

## 3. Test instances

Mirror the production instances, but in pure terms.

```haskell
instance MonadTodoDb TestM where
  findTodo tid = gets (find (\t -> todoId t == tid) . tsTodos)

  insertTodo title = do
    s <- get
    let new = Todo (tsNextId s) title False
    put s { tsTodos = new : tsTodos s, tsNextId = tsNextId s + 1 }
    pure new

  updateTodo todo = modify $ \s ->
    s { tsTodos = map replace (tsTodos s) }
    where
      replace t
        | todoId t == todoId todo = todo
        | otherwise               = t

instance MonadLogger TestM where
  logInfo  msg = modify $ \s -> s { tsLogs = msg : tsLogs s }
  logError msg = modify $ \s -> s { tsLogs = ("[ERR] " <> msg) : tsLogs s }
```

Logs are captured rather than silenced — useful for asserting "did this code path log what it should have?"

---

## 4. Writing assertions

The result of `runTest` is `Either AppError (a, TestState)`. Match on it:

```haskell
-- Happy path
test_createTodo_succeeds :: IO ()
test_createTodo_succeeds = do
  let result = runTest (createTodo "buy milk")
  case result of
    Right (todo, s) -> do
      assertEqual "title" "buy milk" (todoTitle todo)
      assertEqual "not done" False (todoDone todo)
      assertEqual "one todo in state" 1 (length (tsTodos s))
    Left err ->
      assertFailure ("expected success, got: " <> show err)

-- Domain error: empty title
test_createTodo_rejectsEmpty :: IO ()
test_createTodo_rejectsEmpty = do
  let result = runTest (createTodo "   ")
  case result of
    Left (InvalidTitle _) -> pure ()  -- expected
    Left err              -> assertFailure ("wrong error: " <> show err)
    Right _               -> assertFailure "expected failure, got success"

-- State threading: complete a todo
test_completeTodo_marksDone :: IO ()
test_completeTodo_marksDone = do
  let scenario = do
        t <- createTodo "buy milk"
        completeTodo (todoId t)
  case runTest scenario of
    Right (todo, _) -> assertEqual "is done" True (todoDone todo)
    Left err        -> assertFailure (show err)

-- Domain error: completing a nonexistent todo
test_completeTodo_notFound :: IO ()
test_completeTodo_notFound = do
  case runTest (completeTodo 999) of
    Left (NotFound 999) -> pure ()
    other               -> assertFailure ("unexpected: " <> show other)
```

The shape is always: build a scenario, run it, pattern-match the `Either`, assert on the payload or final state.

---

## 5. Injecting infrastructure failures

To test how business code handles infra errors (`DbError`), the test instance can throw on demand. Two approaches.

### Approach A: a fault-injection field in `TestState`

```haskell
data TestState = TestState
  { tsTodos       :: [Todo]
  , tsNextId      :: Int
  , tsLogs        :: [Text]
  , tsFailNextDb  :: Maybe AppError  -- if Just, next DB call throws
  }

-- In the instance:
instance MonadTodoDb TestM where
  insertTodo title = do
    fault <- gets tsFailNextDb
    case fault of
      Just err -> do
        modify $ \s -> s { tsFailNextDb = Nothing }
        throwError err
      Nothing -> do
        s <- get
        let new = Todo (tsNextId s) title False
        put s { tsTodos = new : tsTodos s, tsNextId = tsNextId s + 1 }
        pure new
  -- (similar for findTodo, updateTodo)

-- Use in a test:
test_createTodo_handlesDbFailure :: IO ()
test_createTodo_handlesDbFailure = do
  let s0 = emptyState { tsFailNextDb = Just (DbError "connection lost") }
  case runTestM s0 (createTodo "buy milk") of
    Left (DbError _) -> pure ()
    other            -> assertFailure (show other)
```

### Approach B: a separate "failing" instance

If you only need one failing scenario, write a dedicated newtype:

```haskell
newtype FailingDbM a = FailingDbM (Except AppError a)
  deriving newtype (Functor, Applicative, Monad, MonadError AppError)

instance MonadTodoDb FailingDbM where
  findTodo   _ = throwError (DbError "always fails")
  insertTodo _ = throwError (DbError "always fails")
  updateTodo _ = throwError (DbError "always fails")
```

Approach A is more flexible (lets you sequence "succeeds, succeeds, fails"). Approach B is simpler when one failure mode is enough.

---

## 6. When to use which test monad

The right test monad depends on what the function exercises:

| Function shape | Test monad |
|----------------|------------|
| No state, pure validation chains | `Identity` over `Except AppError` |
| Reads/writes fake state | `StateT TestState (Except AppError)` (the default above) |
| Needs to test concurrency or async | `IO` with real instances + test doubles in `Env` (close to integration test) |
| Calls only one capability, no error handling tested | Direct mock function, skip the monad |

The default `TestM` covers most cases. Reach for `Identity`-based stacks only when there's literally no state — they're slightly faster but the win is marginal.

---

## 7. What you do *not* test in the business layer

Resist the temptation to test these at Layer 2:

- **SQL queries** — that's Layer 1. Test those against a real database (or an in-memory equivalent) in a separate test suite.
- **HTTP routing/serialization** — that's the boundary. Test it with real HTTP requests.
- **The `tryIO` boundary itself** — write one or two integration tests at Layer 1; don't relitigate in business tests.
- **Whether `MonadError` short-circuits** — that's testing the library, not your code.

The business layer is for testing **business rules**: validation logic, lookup behavior, conditional flows, error propagation through composed operations.

---

## 8. Property tests

Polymorphic business functions are excellent targets for QuickCheck/Hedgehog. The state monad gives you a deterministic "world" you can shrink against.

```haskell
prop_createTodoIncreasesCountByOne :: Property
prop_createTodoIncreasesCountByOne = property $ do
  title <- forAll (Gen.text (Range.linear 1 100) Gen.alpha)
  let result = runTest (createTodo title)
  case result of
    Right (_, s) -> length (tsTodos s) === 1
    Left err     -> annotate ("unexpected: " <> show err) >> failure
```

Property tests against the pure `TestM` are fast (no IO) and deterministic (no test database state). This is the regime where this pattern pays off most.

---

## 9. Anti-patterns

```haskell
-- ✗ Using IO in tests when State suffices
test_createTodo = do
  ref <- newIORef []
  let env = Env { envTodoStore = ref }
  result <- runApp env (createTodo "buy milk")
  ...

-- ✓ Use TestM; no IO, no setup teardown
test_createTodo =
  case runTest (createTodo "buy milk") of
    Right (todo, _) -> assertEqual "title" "buy milk" (todoTitle todo)
    Left err        -> assertFailure (show err)
```

```haskell
-- ✗ Mocking by editing production instance
instance MonadTodoDb AppM where
  findTodo tid
    | testMode  = ...
    | otherwise = realDb tid

-- ✓ Use a different monad for tests
```

```haskell
-- ✗ Re-implementing business logic in the test
test_createTodo = do
  -- this doesn't test createTodo, it tests the test
  let title = validateTitle "buy milk"
  ...

-- ✓ Run the actual function
test_createTodo = case runTest (createTodo "buy milk") of ...
```

The principle: test the function the user will actually call, in the simplest monad that supports it.

---

## 10. When the production instance uses a handle

If production gives a capability a handle (for decoration or runtime swap, see [handles-upgrade.md](handles-upgrade.md)), the pure-state test instance above is incompatible with that handle's `IO`-typed fields.

There are three ways out:

1. **Drop the handle for the to-be-faked capability.** Keep handles where they earn their keep (DB with logging/retry), use pure typeclass instances where they don't (email). The pure fake stays straightforward.

2. **Use `IO + IORef` as a test substrate.** Keep all handles. Make `TestM` a `MonadIO` and back fakes with `IORef`s. Tests aren't pure, but the production code path is exercised.

3. **Parameterize the handle by monad** (`Handle m`). The handle can hold `TestM` actions in tests. Most flexible, most type plumbing.

Pick per capability, not per app. See [handles-upgrade.md](handles-upgrade.md) section 9 for the full discussion.
