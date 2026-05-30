# Upgrading to Handles

Companion to [SKILL.md](SKILL.md). This is an **upgrade path** on top of the typeclass approach, not a replacement. Typeclasses stay; handles are added inside their instances.

## When typeclass-only isn't enough

Reach for handles when one or more of these signals appear:

### 1. Cross-cutting decoration

The capability needs to be wrapped with logging, retry, metrics, tracing, or circuit-breaking — and the wrapping logic shouldn't pollute the business layer or the production instance.

```haskell
-- Want: postgresDb + logging + retry + metrics
-- With pure typeclasses you need newtype wrappers + DerivingVia + manual forwarding.
-- With handles: loggingDb (retryingDb 3 (metricsDb (postgresDb conn)))
```

### 2. Multiple implementations of one capability live at once

The app needs a primary and a replica DB, or shards, or A/B implementations. Typeclasses dispatch by *type*; you can only have one instance per `(class, monad)` pair. Handles dispatch by *value*; you can have as many as you want.

```haskell
data Env = Env
  { envPrimaryDb :: TodoDb
  , envReplicaDb :: TodoDb
  , envShardDbs  :: Map Region TodoDb
  }
```

### 3. Runtime swapping via feature flags

The implementation must change at runtime — e.g., "if flag X is on, use the new backend." A handle is a value; swap it. A typeclass instance is fixed at compile time.

### 4. Third-party libraries that hand you a record of functions

Some libraries already model capabilities as records (mock-friendly DB drivers, http-client managers, etc.). Wrapping them in a typeclass is ceremony; using them as-is via a handle is direct.

### 5. Per-request capability variation

Per-tenant DB connections, per-user feature toggles, request-scoped tracing IDs — all naturally expressed as "the handle for this request differs." Typeclass dispatch can't see that; handles can.

If none of these apply, stay with pure typeclasses. Don't add machinery you don't need.

---

## 1. The hybrid shape

The typeclass **interface** stays unchanged. Each capability also gets a corresponding **handle record**. The instance becomes a thin forwarder.

```haskell
-- ─── The typeclass interface (unchanged from SKILL.md) ────

class Monad m => MonadTodoDb m where
  findTodo   :: Int  -> m (Maybe Todo)
  insertTodo :: Text -> m Todo
  updateTodo :: Todo -> m ()

-- ─── The handle: a record mirroring the class ────────────

data TodoDb = TodoDb
  { _findTodo   :: Int  -> IO (Maybe Todo)
  , _insertTodo :: Text -> IO Todo
  , _updateTodo :: Todo -> IO ()
  }

-- ─── Env carries the handle ──────────────────────────────

data Env = Env
  { envTodoDb :: TodoDb
  -- other handles...
  }

-- ─── withDb forwards from Env handle to IO ───────────────

withDb :: (TodoDb -> IO a) -> AppM a
withDb f = asks envTodoDb >>= tryIO . f

-- ─── The instance: thin forwarders ───────────────────────

instance MonadTodoDb AppM where
  findTodo   tid   = withDb (\h -> _findTodo   h tid)
  insertTodo title = withDb (\h -> _insertTodo h title)
  updateTodo todo  = withDb (\h -> _updateTodo h todo)
```

Business code is unchanged. It still calls `findTodo`, `insertTodo`, `updateTodo`. The instance pulls the handle from `Env` and forwards.

---

## 2. Constructing handles

A handle is just a record. Construct one for each implementation.

```haskell
-- Production: real database
postgresDb :: Connection -> TodoDb
postgresDb conn = TodoDb
  { _findTodo   = \tid   -> findTodoSql   conn tid
  , _insertTodo = \title -> insertTodoSql conn title
  , _updateTodo = \todo  -> updateTodoSql conn todo
  }

-- Tests: in-memory
inMemoryDb :: IORef [Todo] -> IORef Int -> TodoDb
inMemoryDb todosRef nextIdRef = TodoDb
  { _findTodo = \tid -> do
      todos <- readIORef todosRef
      pure (find (\t -> todoId t == tid) todos)
  , _insertTodo = \title -> do
      tid <- readIORef nextIdRef
      writeIORef nextIdRef (tid + 1)
      let new = Todo tid title False
      modifyIORef todosRef (new :)
      pure new
  , _updateTodo = \todo -> modifyIORef todosRef $ map $ \t ->
      if todoId t == todoId todo then todo else t
  }
```

Build whichever one fits the environment:

```haskell
prodEnv  conn = Env { envTodoDb = postgresDb conn }
testEnv  ref nid = Env { envTodoDb = inMemoryDb ref nid }
```

---

## 3. Decorators — the main payoff

Decorators wrap a handle and return a new handle. They're just functions on records.

### Logging decorator

```haskell
loggingDb :: Logger -> TodoDb -> TodoDb
loggingDb logger inner = TodoDb
  { _findTodo = \tid -> do
      _logInfo logger ("findTodo " <> T.pack (show tid))
      _findTodo inner tid
  , _insertTodo = \title -> do
      _logInfo logger ("insertTodo " <> title)
      _insertTodo inner title
  , _updateTodo = \todo -> do
      _logInfo logger ("updateTodo " <> T.pack (show (todoId todo)))
      _updateTodo inner todo
  }
```

### Retry decorator

```haskell
retryingDb :: Int -> TodoDb -> TodoDb
retryingDb maxAttempts inner = TodoDb
  { _findTodo   = retry maxAttempts . _findTodo   inner
  , _insertTodo = retry maxAttempts . _insertTodo inner
  , _updateTodo = retry maxAttempts . _updateTodo inner
  }
  where
    retry :: Int -> IO a -> IO a
    retry n action
      | n <= 1    = action
      | otherwise = action `catch` \(_ :: SomeException) -> retry (n - 1) action
```

### Metrics decorator

```haskell
metricsDb :: MetricsClient -> TodoDb -> TodoDb
metricsDb m inner = TodoDb
  { _findTodo = \tid -> timed m "findTodo" (_findTodo inner tid)
  , _insertTodo = \t -> timed m "insertTodo" (_insertTodo inner t)
  , _updateTodo = \t -> timed m "updateTodo" (_updateTodo inner t)
  }
```

### Composition

Decorators compose by function application:

```haskell
mkProductionDb :: Connection -> Logger -> MetricsClient -> TodoDb
mkProductionDb conn logger metrics =
    loggingDb logger
  $ metricsDb metrics
  $ retryingDb 3
  $ postgresDb conn
```

Read inside-out: postgres at the core, retry around it, metrics around that, logging at the outermost layer. Each decorator is independently testable and swappable.

**This is the payoff.** Doing the same with pure typeclasses would require a `newtype` per decorator, `DerivingVia` boilerplate, and instance forwarding for every method. Handles make decoration a one-line function.

---

## 4. Per-request handle variation

Because handles are values, you can construct them per request:

```haskell
-- Choose DB based on tenant
mkTenantDb :: TenantId -> AppM TodoDb
mkTenantDb tid = do
  shards <- asks envShardDbs
  case Map.lookup (shardFor tid) shards of
    Just db -> pure db
    Nothing -> throwError (DbError "unknown shard")

-- Use it for a request-scoped operation
handleRequest :: TenantId -> Text -> AppM Todo
handleRequest tid title = do
  db <- mkTenantDb tid
  -- temporarily use db for this operation
  local (\env -> env { envTodoDb = db }) $ insertTodo title
```

`local` from `MonadReader` swaps the env for a sub-action. The typeclass instance pulls from `envTodoDb`, so swapping it changes which implementation runs — without changing the business code.

---

## 5. The cost: signature duplication

Each capability now exists in three places:

```haskell
-- 1. The class
class Monad m => MonadTodoDb m where
  findTodo   :: Int  -> m (Maybe Todo)
  insertTodo :: Text -> m Todo
  updateTodo :: Todo -> m ()

-- 2. The handle
data TodoDb = TodoDb
  { _findTodo   :: Int  -> IO (Maybe Todo)
  , _insertTodo :: Text -> IO Todo
  , _updateTodo :: Todo -> IO ()
  }

-- 3. The forwarding instance
instance MonadTodoDb AppM where
  findTodo   tid   = withDb (\h -> _findTodo   h tid)
  insertTodo title = withDb (\h -> _insertTodo h title)
  updateTodo todo  = withDb (\h -> _updateTodo h todo)
```

For a capability with ten methods, this is annoying. Mitigations:

- **Accept it.** Most projects do. The duplication is mechanical and local.
- **Generate with Template Haskell.** Libraries like `makeFunctorialField` or hand-rolled TH can derive the handle and forwarding instance from the class. Adds a dependency and some magic; usually not worth it.
- **Skip the record, put functions directly in `Env`.** Loses the grouping ("the DB" as one value) but cuts duplication.

For most codebases, accept the duplication. The flexibility is worth it.

---

## 6. When to keep the typeclass

A reader of the hybrid might wonder: "If I have handles in `Env`, why also have the typeclass?"

Keep the class because:

- **Business signatures stay polymorphic.** `(MonadTodoDb m, MonadError AppError m) => ...` is the signature that lets tests substitute a different `m`. Without the class, business code becomes `AppM`-specific.
- **Documentation.** The class declaration is a clear list of "this is what the capability offers." A record buried in `Env` is harder to discover.
- **Mock with a different monad.** Tests can write `instance MonadTodoDb TestM where ...` and run business code under `TestM` — no need to construct a fake handle.

If the business code is *not* polymorphic (everything is in `AppM`), the typeclass earns its keep less. But then you lose pure-state testing too. Pay the duplication cost for the polymorphism win.

---

## 7. Migration path from pure typeclasses

Adding handles to a typeclass-only codebase is a low-risk refactor:

1. **For each capability**, define the handle record mirroring the class.
2. **Add a field to `Env`** holding the handle.
3. **Rewrite the instance** to forward to the handle (`withX (..._method handle)`).
4. **Build the handle from the existing implementation** — extract the bodies of the old instance methods into handle constructors.
5. **Verify nothing changed for business code** — signatures, behavior, tests should be unchanged.

Once that's mechanical and green, decorators become available:

6. **Identify a cross-cutting concern** that motivated the migration (logging, retry, ...).
7. **Write a decorator function** wrapping `Handle -> Handle`.
8. **Apply it when constructing `Env`** for whichever environments need it.

Now you can add or remove the decorator without touching the rest of the code. That's the value the upgrade buys.

---

## 8. Anti-patterns

```haskell
-- ✗ Removing the typeclass after migrating to handles
-- (Now business code is AppM-specific; pure tests get harder.)
data Env = Env { envTodoDb :: TodoDb }
createTodo :: Text -> AppM Todo
createTodo = ...

-- ✓ Keep both: typeclass for the interface, handle for the implementation
class Monad m => MonadTodoDb m where ...
createTodo :: (MonadTodoDb m, ...) => Text -> m Todo
```

```haskell
-- ✗ Decorating in the instance instead of in the handle
instance MonadTodoDb AppM where
  insertTodo title = do
    logInfo ("inserting " <> title)  -- decoration here
    withDb (\h -> _insertTodo h title)

-- ✓ Decoration is a handle wrapper, not instance logic
loggingDb logger inner = TodoDb { _insertTodo = \t -> logIt t >> _insertTodo inner t, ... }
```

```haskell
-- ✗ Putting decorations inside the production implementation
postgresDb conn = TodoDb
  { _insertTodo = \title -> do
      retry 3 $ do
        logInfo ("..." :: Text)
        insertTodoSql conn title
  , ...
  }

-- ✓ Compose decorators as wrappers
mkDb conn logger = loggingDb logger (retryingDb 3 (postgresDb conn))
```

The principle: handles are *values you compose*. Decorate by wrapping, not by editing.

---

## 9. Tension with pure-fake tests

A handle's fields are `IO`-typed, which conflicts with pure-state test fakes. If you give a capability a handle in production and want to fake it purely in `State` for tests, the types don't match.

**The problem**: You have `TodoDb` with `IO` fields in production, but you want `instance MonadTodoDb TestM` backed by pure `State`, and there's no way to put a `State` action in an `IO`-typed field.

**The key insight**: The typeclass-versus-handle choice is **per-capability, not per-app**. Different capabilities have different needs.

### Three solution strategies

1. **Drop the handle for the capability that needs pure fakes.** Keep handles where decoration matters (DB with logging/retry), use pure typeclass instances where it doesn't (email). The pure fake becomes straightforward. This is usually the right answer.

2. **Use `IO + IORef` as a uniform test substrate.** Keep all handles. Make `TestM` a `MonadIO` and back fakes with `IORef`s. Tests aren't pure, but the production handle paths stay exercised. Good for integration-style tests.

3. **Parameterize the handle by monad** (`Handle m` instead of `Handle`). The handle can hold `AppM` actions in production and `TestM` actions in tests. Most flexible, most type plumbing. Reach for this only when you need both decoration AND pure tests of the same capability.

### How to choose

| You want | Pick |
|----------|------|
| Pure tests AND don't need to decorate this capability | Drop the handle for this capability (option 1) |
| Decoration in production AND happy with `IO`-based tests | `IORef`-backed fakes (option 2) |
| Decoration in production AND pure tests of the SAME capability | Polymorphic handle `Handle m` (option 3) |

**A frequent mistake**: Deciding once for the whole app. Mixed answers across capabilities are normal and correct.
