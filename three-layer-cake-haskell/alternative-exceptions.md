# Alternative: Explicit Either Returns for Domain Errors

Companion to [SKILL.md](SKILL.md). This is an alternative track where domain errors are returned as explicit `Either DomainError a` values instead of flowing through `MonadError`.

**Note**: The default track (SKILL.md) now uses a hand-written `MonadError` instance over `ReaderT Env IO`—**not** `ExceptT`. That implementation composes cleanly with `async`/`concurrently` because `throwError` raises a real `IO` exception. If you were considering this alternative because `ExceptT` broke your concurrent code, check the default track first—it likely already solves that problem.

## When to suggest this variant

Reach for this when any of these apply:

- **The user prefers explicit `Either` in return types.** Some teams find `MonadError`'s hidden control flow uncomfortable and want the "can-fail" nature visible in every signature.
- **The codebase already uses exceptions extensively for infra.** Adding `MonadError` for domain errors creates a second channel; this variant consolidates infra to exceptions and domain to explicit returns.
- **Heavy use of streaming libraries** (`conduit`, `pipes`) where the inner monad needs to be plain `IO` and explicit `Either` threading is preferred over typeclass-based error handling.

Do **not** switch to this variant just because exceptions feel more familiar. The default track's `MonadError` has real benefits: polymorphic business signatures, cleaner do-blocks, and easier pure tests. Switch when there's a concrete preference for explicit `Either` returns.

---

## 1. The shape

`AppM` is now just `ReaderT Env IO`. No `ExceptT`. Infra errors are exceptions; domain errors are `Either DomainError a` return values.

```haskell
{-# LANGUAGE GeneralizedNewtypeDeriving #-}
{-# LANGUAGE DerivingStrategies #-}
{-# LANGUAGE OverloadedStrings #-}

import Control.Monad.Reader
import Control.Exception (Exception)
import UnliftIO.Exception (throwIO, catch)
import Data.Text (Text)

-- ─── Layer 3: Pure core (mostly unchanged) ────────────────

data Todo = Todo { todoId :: Int, todoTitle :: Text, todoDone :: Bool }

-- Domain errors: only the things business code needs to branch on.
data DomainError
  = NotFound Int
  | InvalidTitle Text
  | AlreadyCompleted Int
  deriving Show

-- Infra errors: thrown as exceptions. Caught only at the boundary.
data InfraException
  = DbException Text
  | NetworkException Text
  deriving Show

instance Exception InfraException

-- ─── Layer 2: app monad — just ReaderT IO ─────────────────

data Env = Env { envConn :: Conn }
data Conn = Conn

newtype AppM a = AppM { unAppM :: ReaderT Env IO a }
  deriving newtype (Functor, Applicative, Monad, MonadReader Env, MonadIO)

runApp :: Env -> AppM a -> IO a
runApp env = flip runReaderT env . unAppM
```

Notice: no `ExceptT`, no `MonadError`. `runApp` returns `IO a` directly. Exceptions propagate through.

---

## 2. Capability classes

Methods still don't return `Either` for infra failures — those throw. But they **do** return `Either DomainError` when domain failure is a normal outcome of the operation.

```haskell
class Monad m => MonadTodoDb m where
  findTodo   :: Int  -> m (Maybe Todo)
  insertTodo :: Text -> m Todo
  updateTodo :: Todo -> m ()
```

Wait — these look identical to the default track. They are, at the *type* level. The difference is the contract:

- **Default track**: an infra failure inside `insertTodo` calls `throwError (DbError ...)` and the calling do-block short-circuits via `ExceptT`.
- **This track**: an infra failure inside `insertTodo` calls `throwIO (DbException ...)` and the exception unwinds the call stack until something catches it.

The instance does the throwing; the business code looks the same either way. The difference is visible at the boundary.

---

## 3. The withDb helper, exception-flavored

```haskell
withDb :: (Conn -> IO a) -> AppM a
withDb f = do
  c <- asks envConn
  liftIO (f c `catch` \(e :: SomeException) ->
           throwIO (DbException (T.pack (show e))))
```

Or, more cleanly, let the underlying `IO` exception propagate as-is and only wrap if you need a uniform type:

```haskell
withDb :: (Conn -> IO a) -> AppM a
withDb f = asks envConn >>= liftIO . f
```

If the driver throws `SqlError`, it just propagates. The boundary catches `SomeException` and decides what to do. The point of wrapping with `DbException` is to have a single type to match in the handler — pick based on how granular you want the error channel.

---

## 4. Business logic with explicit `Either` for domain

Business operations that can fail with a domain error return `Either DomainError a`:

```haskell
createTodo
  :: (MonadTodoDb m)
  => Text -> m (Either DomainError Todo)
createTodo raw = case validateTitle raw of
  Left err    -> pure (Left err)
  Right title -> Right <$> insertTodo title

completeTodo
  :: (MonadTodoDb m)
  => Int -> m (Either DomainError Todo)
completeTodo tid = do
  mTodo <- findTodo tid
  case mTodo of
    Nothing -> pure (Left (NotFound tid))
    Just todo -> case markDone todo of
      Left err   -> pure (Left err)
      Right done -> do
        updateTodo done
        pure (Right done)
```

Notice the noise. This is the cost of giving up `MonadError`: explicit `case` matches at every fallible step. You can mitigate with `ExceptT` locally:

```haskell
completeTodo tid = runExceptT $ do
  mTodo <- lift (findTodo tid)
  todo  <- maybe (throwError (NotFound tid)) pure mTodo
  done  <- liftEither (markDone todo)
  lift (updateTodo done)
  pure done
```

This is the pragmatic middle ground: `ExceptT` *locally* inside a function, but the function's signature exposes `m (Either DomainError a)` — so the outer code (especially concurrent code) deals with plain `IO`.

---

## 5. Handling at the HTTP boundary

The handler catches infra exceptions and pattern-matches on domain `Either`:

```haskell
handleCreateTodo :: Env -> Text -> IO Response
handleCreateTodo env rawTitle = do
  result <- (Right <$> runApp env (createTodo rawTitle))
              `catch` (\(e :: InfraException) -> pure (Left e))
  case result of
    Left (DbException msg) -> do
      logErrorIO msg
      respond 500 (encode ("internal error" :: Text))
    Left (NetworkException msg) -> do
      logErrorIO msg
      respond 502 (encode ("upstream error" :: Text))
    Right (Left (InvalidTitle msg))     -> respond 400 (encode msg)
    Right (Left (NotFound tid))         -> respond 404 (encode tid)
    Right (Left (AlreadyCompleted tid)) -> respond 409 (encode tid)
    Right (Right todo)                  -> respond 201 (encode todo)
```

Note the nesting: `Either InfraException (Either DomainError Todo)`. Three outcomes — infra failure, domain failure, success — flow through two different mechanisms. This is the boundary tax.

Factor it out with a helper:

```haskell
runHandler :: Env -> AppM (Either DomainError a) -> (a -> IO Response) -> IO Response
runHandler env action onSuccess = do
  result <- try (runApp env action)
  case result of
    Left (DbException msg) -> do
      logErrorIO msg
      respond 500 (encode ("internal error" :: Text))
    Left (NetworkException msg) -> do
      logErrorIO msg
      respond 502 (encode ("upstream" :: Text))
    Right (Left dErr)  -> domainErrorToResponse dErr
    Right (Right a)    -> onSuccess a

domainErrorToResponse :: DomainError -> IO Response
domainErrorToResponse = \case
  InvalidTitle msg     -> respond 400 (encode msg)
  NotFound tid         -> respond 404 (encode tid)
  AlreadyCompleted tid -> respond 409 (encode tid)
```

Handlers become one line:

```haskell
handleCreateTodo env rawTitle =
  runHandler env (createTodo rawTitle) (respond 201 . encode)
```

---

## 6. Concurrency note

**Historical context**: This file was written when the default track used `ExceptT` over `IO`, which composed badly with `async`/`concurrently`. The default track now uses a hand-written `MonadError` instance backed by real `IO` exceptions, which composes cleanly.

With the updated default track, concurrent code works naturally:

```haskell
-- Default track (now works because throwError raises a real exception):
doBoth :: AppM (Todo, Todo)
doBoth = do
  env <- ask
  liftIO $ concurrently
    (runApp' env (createTodo "a"))
    (runApp' env (createTodo "b"))
  where
    runApp' env action = do
      result <- runApp env action
      either (throwIO . AppException) pure result
```

The main remaining reason for this alternative is **preference for explicit `Either` returns** (making the "can-fail" nature visible in every business function signature) rather than concurrency composition.

---

## 7. Testing in the alternative track

The exception variant complicates testing slightly because instances now perform IO directly (no `MonadError` to short-circuit cleanly in a pure stack).

Options:

1. **Test domain logic only** — call business functions that return `Either DomainError a`, assert on the `Either`. Use a `ReaderT TestEnv IO` with fake instances.
2. **Use `IO` in tests** — accept the small cost; build a fake `Env` with `IORef`-backed instances; assert on results.

```haskell
data TestEnv = TestEnv { teTodos :: IORef [Todo], teNextId :: IORef Int }

newtype TestM a = TestM (ReaderT TestEnv IO a)
  deriving newtype (Functor, Applicative, Monad, MonadReader TestEnv, MonadIO)

instance MonadTodoDb TestM where
  findTodo tid = do
    ref <- asks teTodos
    todos <- liftIO (readIORef ref)
    pure (find (\t -> todoId t == tid) todos)
  -- ...

runTest :: TestM a -> IO a
runTest (TestM m) = do
  env <- TestEnv <$> newIORef [] <*> newIORef 1
  runReaderT m env
```

For domain assertions:
```haskell
test_createTodo = do
  result <- runTest (createTodo "buy milk")
  case result of
    Right todo -> assertEqual "title" "buy milk" (todoTitle todo)
    Left err   -> assertFailure (show err)
```

For infra failure injection: throw exceptions inside the instance based on an `IORef` flag, then `try` in the test.

Pure tests are no longer free — that's the tax for this track.

---

## 8. Tradeoff summary

| Concern | Default (`MonadError`) | Alternative (exceptions) |
|---------|------------------------|--------------------------|
| Sequential business code | Clean do-blocks | Clean if local `ExceptT` is used |
| Concurrent composition | Painful (`ExceptT` doesn't compose) | Natural (`async` works) |
| Signature documents failure | Yes (`MonadError AppError m`) | Partial (only domain via return type) |
| Boundary complexity | Pattern-match one `Either` | Catch exception + match `Either` |
| Test purity | Easy (`Identity`/`State`) | Requires `IO` stack |
| Async-exception safety | Yes (via `UnliftIO`) | Yes (use `UnliftIO` for `catch`) |
| Mental model | One channel | Two channels (exception + return) |

The default is better for most apps. Switch to the alternative only when your team strongly prefers explicit `Either` returns over `MonadError`'s hidden control flow.

---

## 9. Don't mix mid-stream

If a codebase has been using the default track, don't introduce exceptions for "just this one thing." The mixed style is the worst of both worlds: readers can't tell which functions throw, error handling becomes a guessing game, and refactors break in subtle ways.

Pick a track per **bounded context** (subsystem). Within that context, be consistent. Different contexts can use different tracks — that's fine as long as the boundary between them is explicit.
