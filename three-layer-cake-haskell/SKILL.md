---
name: three-layer-cake-haskell
description: Architecting Haskell applications using the three-layer cake pattern with typeclass-based capabilities, MonadError for domain errors, and IO exceptions for infrastructure errors. Use this skill when the user explicitly asks for the three-layer cake, the ReaderT pattern, or capability typeclasses. Also consult this skill proactively when the user is discussing how to structure, organize, or architect a non-trivial Haskell application — especially when they mention layering, separating IO from business logic, managing effects, error handling across an app, testable business logic, swappable infrastructure, rolling back side effects on failure, sagas, or compensating transactions. Do not push this pattern on simple scripts, one-shot tools, or prototypes; do not refuse to apply it if the user asks for it by name regardless of context.
---

# Three-Layer Cake (Haskell)

The three-layer cake separates a Haskell application into:

1. **Pure core** — domain types and pure functions
2. **Business logic** — composed in an app monad with capability typeclasses, using `MonadError` for domain errors
3. **Imperative shell** — `IO`, instances that wrap raw resources, error boundary that catches infra exceptions

The default track in this skill: typeclasses for capabilities, `withDb`-style helpers in instances, `MonadError AppError` as the *interface* (not `ExceptT`), and exceptions in `IO` for infra errors caught at the boundary.

**Important**: The app monad is `ReaderT Env IO` with a hand-written `MonadError` instance—**not** `ExceptT` over `IO`. The well-known advice "don't use `ExceptT` over `IO`" targets that specific implementation, not the `MonadError` interface. Stacking `ExceptT` on `IO` causes misleading signatures, runtime cost, and broken composition with `concurrently`. The cake keeps the interface and drops that implementation.

---

## 1. When not to proactively apply this pattern

This pattern earns its keep when:
- The app has non-trivial domain rules worth isolating from infra
- Infrastructure may be swapped (Postgres ↔ SQLite, real services ↔ fakes for tests)
- Multiple developers need clear seams between layers
- Business logic needs to be testable without IO

Do **not** proactively suggest this pattern for:
- One-off scripts and CLI utilities
- Glue code and shell-replacement tools
- Tutorials, prototypes, single-file demos
- Learning exercises where the user is exploring a different concept

**Critical override**: if the user explicitly asks for the three-layer cake, the ReaderT pattern, capability typeclasses, or describes a setup that matches this pattern, apply it without protest. Do not suggest a "simpler" alternative unprompted. The user knows their context better than you do.

---

## 2. The three layers

```
┌──────────────────────────────────────────────────────────┐
│ Layer 1: Imperative shell (IO)                           │
│   - Resources: DB connections, HTTP clients, files       │
│   - Typeclass instances that wrap raw IO                 │
│   - tryIO boundary: converts exceptions → AppError       │
│   - HTTP handlers: runApp, map errors to responses       │
├──────────────────────────────────────────────────────────┤
│ Layer 2: Business logic (AppM)                           │
│   - Polymorphic over (MonadX m, MonadError AppError m)   │
│   - Composes capabilities in flat do-blocks              │
│   - Uses throwError / liftEither' for domain errors      │
│   - Never sees IOException; never calls liftIO directly  │
├──────────────────────────────────────────────────────────┤
│ Layer 3: Pure core                                       │
│   - Domain types (Todo, User, Order)                     │
│   - Pure validation and rules → Either AppError a        │
│   - No IO, no monad transformers                         │
└──────────────────────────────────────────────────────────┘
```

**Error model**:
- **Domain errors** (NotFound, InvalidInput, BusinessRuleViolation) — modeled as constructors of `AppError`, raised via `throwError`.
- **Infra errors** (connection lost, SQL constraint, timeout) — raised as `IO` exceptions by the underlying library, caught by `tryIO` at the boundary, converted to an `AppError` variant (typically `DbError` or `InfraError`).

This split is the whole point. Business code only branches on domain outcomes. Infra failures short-circuit silently via the error monad.

---

## 3. The canonical skeleton

A runnable starting point. Every three-layer-cake app looks like this.

```haskell
{-# LANGUAGE GeneralizedNewtypeDeriving #-}
{-# LANGUAGE DerivingStrategies         #-}
{-# LANGUAGE FlexibleContexts           #-}
{-# LANGUAGE LambdaCase                 #-}
{-# LANGUAGE OverloadedStrings          #-}
{-# LANGUAGE ScopedTypeVariables        #-}

module App where

import Control.Exception      (Exception, SomeException)
import Control.Monad.Except   (MonadError (..))
import Control.Monad.Reader
import Data.Text              (Text)
import qualified Data.Text as T
import UnliftIO.Exception     (catch, throwIO, try)

-- ─── Layer 3: Pure core ───────────────────────────────────

data Todo = Todo
  { todoId    :: Int
  , todoTitle :: Text
  , todoDone  :: Bool
  } deriving Show

data AppError
  = NotFound Int
  | InvalidTitle Text
  | AlreadyCompleted Int
  | DbError Text
  deriving Show

validateTitle :: Text -> Either AppError Text
validateTitle t
  | T.null (T.strip t) = Left (InvalidTitle "empty title")
  | T.length t > 200   = Left (InvalidTitle "title too long")
  | otherwise          = Right (T.strip t)

markDone :: Todo -> Either AppError Todo
markDone todo
  | todoDone todo = Left (AlreadyCompleted (todoId todo))
  | otherwise     = Right todo { todoDone = True }

-- ─── Layer 1: Capability classes ──────────────────────────

class Monad m => MonadTodoDb m where
  findTodo   :: Int  -> m (Maybe Todo)
  insertTodo :: Text -> m Todo
  updateTodo :: Todo -> m ()

class Monad m => MonadLogger m where
  logInfo  :: Text -> m ()
  logError :: Text -> m ()

-- ─── Layer 2: The app monad ───────────────────────────────

-- Env carries resources (connection pools, refs, configs).
-- Assume `Conn` is whatever your DB library hands you.
data Env = Env
  { envConn :: Conn  -- placeholder; real type from your driver
  }
data Conn = Conn  -- stub; replace with postgresql-simple Connection etc.

-- The wrapper that carries domain errors as real IO exceptions.
newtype AppException = AppException AppError
  deriving Show
instance Exception AppException

-- ReaderT over IO. No ExceptT. Domain errors travel as IO exceptions
-- wrapped in AppException; MonadError is the interface, not the transformer.
newtype AppM a = AppM { unAppM :: ReaderT Env IO a }
  deriving newtype
    ( Functor, Applicative, Monad
    , MonadReader Env
    , MonadIO
    )

-- The hand-written MonadError instance: throwError raises AppException,
-- catchError catches only that wrapper.
instance MonadError AppError AppM where
  throwError = AppM . liftIO . throwIO . AppException
  catchError (AppM action) handler =
    AppM $ ReaderT $ \env ->
      runReaderT action env
        `catch` \(AppException e) -> runReaderT (unAppM (handler e)) env

-- The one place the wrapper is unwound: convert back to Either at the boundary.
runApp :: Env -> AppM a -> IO (Either AppError a)
runApp env (AppM action) =
  (Right <$> runReaderT action env)
    `catch` \(AppException e) -> pure (Left e)

-- ─── The IO boundary ──────────────────────────────────────

-- Catches synchronous exceptions; lets async exceptions through.
-- `try` from UnliftIO.Exception, NOT Control.Exception.
tryIO :: IO a -> AppM a
tryIO action = liftIO (try action) >>= \case
  Right a                   -> pure a
  Left (e :: SomeException) -> throwError (DbError (T.pack (show e)))

-- The withX helper: pull a resource from Env, run IO, lift errors.
-- One per capability. Each capability's instance methods become one-liners.
withDb :: (Conn -> IO a) -> AppM a
withDb f = asks envConn >>= tryIO . f

-- ─── Layer 1: Production instances ────────────────────────

-- Assume these exist somewhere (postgresql-simple, sqlite-simple, etc.):
--   findTodoSql   :: Conn -> Int  -> IO (Maybe Todo)
--   insertTodoSql :: Conn -> Text -> IO Todo
--   updateTodoSql :: Conn -> Todo -> IO ()

instance MonadTodoDb AppM where
  findTodo   tid   = withDb $ \c -> findTodoSql   c tid
  insertTodo title = withDb $ \c -> insertTodoSql c title
  updateTodo todo  = withDb $ \c -> updateTodoSql c todo

instance MonadLogger AppM where
  logInfo  msg = liftIO (putStrLn ("[INFO]  " ++ T.unpack msg))
  logError msg = liftIO (putStrLn ("[ERROR] " ++ T.unpack msg))

-- ─── Layer 2: Business logic ──────────────────────────────

liftEither' :: MonadError AppError m => Either AppError a -> m a
liftEither' = either throwError pure

createTodo
  :: (MonadTodoDb m, MonadLogger m, MonadError AppError m)
  => Text -> m Todo
createTodo raw = do
  title <- liftEither' (validateTitle raw)
  logInfo ("creating: " <> title)
  insertTodo title

completeTodo
  :: (MonadTodoDb m, MonadLogger m, MonadError AppError m)
  => Int -> m Todo
completeTodo tid = do
  mTodo <- findTodo tid
  todo  <- maybe (throwError (NotFound tid)) pure mTodo
  done  <- liftEither' (markDone todo)
  updateTodo done
  logInfo ("completed: " <> T.pack (show tid))
  pure done

-- Stubs to keep this snippet compilable:
findTodoSql   :: Conn -> Int  -> IO (Maybe Todo); findTodoSql   _ _ = pure Nothing
insertTodoSql :: Conn -> Text -> IO Todo;         insertTodoSql _ t = pure (Todo 1 t False)
updateTodoSql :: Conn -> Todo -> IO ();           updateTodoSql _ _ = pure ()
```

Hand this skeleton to the user as a starting point and tailor `Env`, `AppError`, and the capability classes to their domain.

---

## 4. The error handling contract

Two kinds of errors, two channels, one rule.

| Kind | Examples | Where it lives | How it flows |
|------|----------|----------------|--------------|
| **Domain** | NotFound, InvalidInput, RuleViolated | Constructors of `AppError` | `throwError` / `liftEither'`; short-circuits the `do`-block via the hand-written `MonadError` instance |
| **Infra** | Connection lost, timeout, SQL error | `IO` exceptions from the driver | Raised by the library, caught by `tryIO`, converted to `AppError` (usually `DbError`) |

**The rule**: business code never sees `IOException`. Every `IO` call from a capability instance goes through `tryIO` (or `withDb`, which folds `tryIO` in). The boundary is the instance, not the business function.

There's a third category: **domain outcomes**. When an alternative result is a normal flow the caller will branch on (not a failure), model it as a return-type sum, not via `MonadError`. See section 4a below.

### Why `UnliftIO.Exception`

```haskell
import UnliftIO.Exception (try)        -- ✓ skips async exceptions
-- import Control.Exception (try)      -- ✗ catches ThreadKilled, UserInterrupt
```

`Control.Exception.try` catches *everything* including async exceptions (`ThreadKilled`, `UserInterrupt`, `System.Timeout`). You almost never want to swallow those — they signal "shut down now." `UnliftIO.Exception.try` only catches synchronous exceptions. Always use it.

### Mapping infra errors more precisely

The skeleton converts every exception to `DbError (show e)`. In production, classify:

```haskell
tryIO :: IO a -> AppM a
tryIO action = do
  result <- liftIO (try action)
  case result of
    Right a -> pure a
    Left (e :: SomeException) -> throwError (classify e)
  where
    classify e
      | Just (SqlError code _ msg _ _) <- fromException e =
          DbError ("SQL " <> T.pack (show code) <> ": " <> T.pack (show msg))
      | Just (ioe :: IOException) <- fromException e =
          DbError ("IO: " <> T.pack (show ioe))
      | otherwise =
          DbError ("unknown: " <> T.pack (show e))
```

`SqlError` here is from `postgresql-simple`; substitute the equivalent for whatever driver is in use. The point is: log the structured information you'll need to debug, then map to a domain-meaningful `AppError`.

### 4a. Domain outcomes

When a command's "failure" is actually part of the state machine (an alternative the caller will branch on), it shouldn't be an error at all. Model it as a sum type and return it. Reserve `MonadError` for things that really are failures.

**Example**: `git push` may be rejected because the remote moved. That's not an error—it's a cue to fetch-rebase-push. But if the network died, that *is* an infra error.

```haskell
data PushOutcome   = Pushed | NeedsRebase
data RebaseOutcome = Rebased | Conflicted [Conflict]

class Monad m => MonadGit m where
  gitPush        :: m PushOutcome
  gitFetch       :: m ()
  gitRebase      :: Text -> m RebaseOutcome
  gitAbortRebase :: m ()
```

The instance classifies exit codes:

```haskell
instance MonadGit AppM where
  gitPush = withGit $ \cwd -> do
    (exit, _, err) <- runGit cwd ["push"]
    case exit of
      ExitSuccess -> pure Pushed
      ExitFailure _ | looksLikeNonFastForward err -> pure NeedsRebase
      ExitFailure n -> throwError (GitError ("push failed: " <> err))
```

Business code becomes a clean state machine:

```haskell
push :: (MonadGit m, MonadError AppError m) => m ()
push = gitPush >>= \case
  Pushed -> pure ()
  NeedsRebase -> do
    gitFetch
    gitRebase "origin/HEAD" >>= \case
      Rebased -> gitPush >>= expectPushed
      Conflicted conflicts -> do
        gitAbortRebase
        throwError (PushConflict conflicts)
  where
    expectPushed Pushed      = pure ()
    expectPushed NeedsRebase = throwError (UnexpectedRebase "after rebase")
```

**The decision**: if the caller will routinely write code for both branches, the alternative belongs in the return type. If the caller treats one branch as a rare exception, the error channel is fine.

| Semantics | Best fit |
|-----------|----------|
| Binary success/failure; failure is rare and means something's wrong | `m ()` that throws via `MonadError` |
| Multiple meaningful results; caller will routinely branch | `m PushOutcome` with explicit constructors |
| Same operation used in both modes across the codebase | Two methods: strict (`gitPush :: m ()`) and permissive (`tryGitPush :: m PushOutcome`) |

---

## 5. Writing business logic

The business layer is where the pattern pays off. Functions read like straight-line code, even though three different failure sources (validation, lookups, business rules) compose through them.

### Polymorphic signatures

Business functions are **polymorphic in `m`**, constrained by exactly the capabilities they use:

```haskell
createTodo
  :: (MonadTodoDb m, MonadLogger m, MonadError AppError m)
  => Text -> m Todo
```

This is not cosmetic. It enables:
- **Test substitution**: run the same function under a `TestM` that fakes `MonadTodoDb` in `State`.
- **Documentation**: the signature lists every effect the function can perform.
- **Refactoring safety**: removing a capability constraint is a compile-time check that the function no longer uses it.

### Pure validation + capability calls + lookups

```haskell
-- ✓ Clean: pure validation, capability call, lookup-or-throw
createTodo raw = do
  title <- liftEither' (validateTitle raw)   -- pure → AppError
  insertTodo title                            -- capability → AppError on infra failure

completeTodo tid = do
  mTodo <- findTodo tid                       -- capability
  todo  <- maybe (throwError (NotFound tid)) pure mTodo  -- lookup → AppError
  done  <- liftEither' (markDone todo)        -- pure rule → AppError
  updateTodo done                             -- capability
  pure done
```

The `liftEither'` helper bridges Layer 3 (pure `Either`) into Layer 2 (the app monad). It's three lines and worth defining in every project:

```haskell
liftEither' :: MonadError AppError m => Either AppError a -> m a
liftEither' = either throwError pure
```

### Anti-patterns

```haskell
-- ✗ Don't unwrap Either manually — that's what MonadError is for
createTodo raw = do
  case validateTitle raw of
    Left err    -> throwError err
    Right title -> insertTodo title

-- ✓ Use liftEither'
createTodo raw = liftEither' (validateTitle raw) >>= insertTodo
```

```haskell
-- ✗ Don't call liftIO in business code
completeTodo tid = do
  liftIO (putStrLn ("completing " ++ show tid))  -- belongs behind MonadLogger
  ...

-- ✓ Use a capability
completeTodo tid = do
  logInfo ("completing " <> T.pack (show tid))
  ...
```

```haskell
-- ✗ Don't catch exceptions in business code
completeTodo tid = do
  result <- liftIO (try (someAction tid))
  ...

-- ✓ Catch at the instance boundary via tryIO / withDb
```

```haskell
-- ✗ Don't return Either from capability methods
class MonadTodoDb m where
  findTodo :: Int -> m (Either AppError (Maybe Todo))

-- ✓ Use MonadError; the monad already carries failures
class Monad m => MonadTodoDb m where
  findTodo :: Int -> m (Maybe Todo)  -- throws via MonadError on infra failure
```

The last one is the most common regression. The whole point of `MonadError` is to avoid manual `Either` plumbing. Returning `Either` from capability methods reintroduces it.

---

## 6. Adding a new capability

The recipe is mechanical. Use it whenever the business layer needs a new kind of effect.

**Step 1: Define the class** (Layer 1 interface)

```haskell
class Monad m => MonadEmail m where
  sendEmail :: Text -> Text -> Text -> m ()  -- to, subject, body
```

**Step 2: Add the resource to `Env`** (if needed)

```haskell
data Env = Env
  { envConn       :: Conn
  , envSmtpClient :: SmtpClient  -- new
  }
```

**Step 3: Write a `withX` helper**

```haskell
withSmtp :: (SmtpClient -> IO a) -> AppM a
withSmtp f = asks envSmtpClient >>= tryIO . f
```

**Step 4: Write the production instance**

```haskell
instance MonadEmail AppM where
  sendEmail to subject body =
    withSmtp $ \client -> sendEmailSmtp client to subject body
```

**Step 5: Use it in business code**

```haskell
notifyTodoCompleted
  :: (MonadTodoDb m, MonadEmail m, MonadError AppError m)
  => Int -> m ()
notifyTodoCompleted tid = do
  mTodo <- findTodo tid
  todo  <- maybe (throwError (NotFound tid)) pure mTodo
  sendEmail "user@example.com" "Done!" (todoTitle todo)
```

That's the whole loop. Add a capability, add a `withX`, add an instance, use it.

---

## 7. Handling at the HTTP boundary

The boundary between Layer 1 (IO) and the outside world is where `AppError` becomes an HTTP response. The shape is framework-independent: `runApp`, pattern-match, map.

```haskell
-- Generic handler shape; adapt to your framework.
handleCreateTodo :: Env -> Text -> IO Response
handleCreateTodo env rawTitle = do
  result <- runApp env (createTodo rawTitle)
  case result of
    Right todo                  -> respond 201 (encode todo)
    Left (InvalidTitle msg)     -> respond 400 (encode msg)
    Left (NotFound tid)         -> respond 404 (encode tid)
    Left (AlreadyCompleted tid) -> respond 409 (encode tid)
    Left (DbError msg) -> do
      -- Log infra details internally; do NOT leak to client.
      logErrorIO msg
      respond 500 (encode ("internal error" :: Text))
```

Two rules at this boundary:

1. **Domain errors map to specific status codes.** The client gets a meaningful response and possibly the error payload.
2. **Infra errors get logged internally and become a generic 5xx.** Never return `DbError` details to clients — they may contain SQL, connection strings, or stack traces.

```haskell
-- ✗ Don't leak infra details
Left (DbError msg) -> respond 500 (encode msg)  -- exposes internals

-- ✓ Log and return generic
Left (DbError msg) -> do
  logErrorIO msg
  respond 500 (encode ("internal error" :: Text))
```

The mapping function can be factored out:

```haskell
errorToResponse :: AppError -> IO Response
errorToResponse = \case
  InvalidTitle msg     -> respond 400 (encode msg)
  NotFound tid         -> respond 404 (encode tid)
  AlreadyCompleted tid -> respond 409 (encode tid)
  DbError msg          -> logErrorIO msg >> respond 500 (encode ("internal error" :: Text))
```

Now every handler is two lines:

```haskell
handleCreateTodo env rawTitle = do
  result <- runApp env (createTodo rawTitle)
  either errorToResponse (respond 201 . encode) result
```

---

## 8. Common pitfalls

Patterns to flag when reviewing user code. Each is a regression of the principles above.

### `liftIO` in business code
Business functions should call capabilities, not `IO` directly. A `liftIO` in Layer 2 is a missing capability.

### DB types in business signatures
`Connection`, `Row`, `Statement` — these are Layer 1 types. If they appear in Layer 2 signatures, the abstraction is leaking. Convert to domain types in the capability instance.

### `Either AppError` returns from capability methods
Reintroduces manual unwrapping. Use `MonadError` instead. (See section 5 anti-patterns.)

Note: this rule targets `Either AppError`-style wrapped returns. A capability returning a domain-outcome sum type like `m PushOutcome` (where `PushOutcome = Pushed | NeedsRebase`) is correct and useful. See section 4a.

### Using `MonadError` for what isn't an error
When a command's alternative outcome is part of the state machine (the caller will routinely branch on it), modeling it as a thrown error hides the alternative path:

```haskell
-- Forces every caller to catchError just to handle the normal alternative (bad)
gitPush :: m ()    -- throws PushRejected when remote moved

-- The alternative is a value the caller pattern-matches on (good)
gitPush :: m PushOutcome
data PushOutcome = Pushed | NeedsRebase
```

See section 4a for the full discussion.

### Catching exceptions inside business code
The boundary is the instance method, not the business function. If business code is catching exceptions, the wrong layer is doing the work.

### `Other Text` escape hatch in `AppError`
Tempting, becomes 80% of errors over time. Force specific constructors. If genuinely unclassifiable, log the details and use a narrow `Internal Text` variant — and treat its appearance as a bug.

### Class-per-trivial-effect
`MonadFileSystem`, `MonadClock`, `MonadRandom` for one-off uses is ceremony. Typeclass effects that have one production instance, no fake needed for tests, and no cross-cutting concerns — just use `liftIO` once and move on. The pattern is for capabilities that *benefit* from polymorphism.

### Mixing the two error tracks
If using `MonadError AppError` (the default), don't *also* throw custom exception types from business code. Pick one channel per error and stick with it.

### Newtype `Conn` everywhere
The `Env` should hold the driver's actual connection type. Don't wrap it unless you need to.

### Multi-step operations with no rollback strategy
A business function that performs two or more side effects (DB writes, external calls) without a rollback plan leaves the system inconsistent on partial failure. `MonadError` short-circuits the computation; it does **not** undo effects already committed to IO. If the user writes code like `doStepA >> doStepB` where partial completion would corrupt state, point them at [compensations.md](compensations.md) for the four approaches: DB transactions, bracket, sagas, and outbox.

---

## 9. Where to go next

Companion files in this skill, each focused on a single concern:

- **[testing.md](testing.md)** — How to test business logic without IO. Pure `TestM` over `State` and `Except`; mock instances; assertion patterns. Read this whenever the user asks how to test, or wants to verify the pattern's testability claim.

- **[alternative-exceptions.md](alternative-exceptions.md)** — An alternative variant where domain errors are returned as `Either DomainError` values (not via `MonadError`) while infra errors remain exceptions. Consider this if the user prefers explicit `Either` in return types and finds `MonadError`'s hidden control flow uncomfortable. The default track (this file) uses `MonadError` with a hand-written instance that still composes cleanly with `async`/`concurrently`.

- **[handles-upgrade.md](handles-upgrade.md)** — When and how to add handles (records of functions) on top of the typeclass approach. Read this when typeclass instances need to be decorated (logging, retry, metrics), when multiple implementations of one capability must live at once, or when runtime swapping is required. This is an upgrade path, not a replacement — the typeclasses stay. Also covers the tension between handle's `IO`-typed fields and pure-state test fakes, with three solution strategies.

- **[compensations.md](compensations.md)** — Undoing side effects when a multi-step operation fails partway. Covers DB transactions, bracket, sagas (compensating actions), and the outbox pattern. Read this when the user asks how to roll back changes on error, mentions sagas or compensating transactions, or describes a multi-step operation where partial completion would corrupt state.

- **[starter-template.hs](starter-template.hs)** — Copy-pasteable skeleton, minimal but complete. Hand this to the user when they want to start a new project with the pattern.

When applying the skill, default to the patterns in this main file. Reach for the alternatives only when the signals in their respective files appear.
