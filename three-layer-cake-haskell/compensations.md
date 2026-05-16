# Compensating Actions and Transactions

Companion to [SKILL.md](SKILL.md). Read this when the user asks how to undo side effects on failure, mentions sagas or compensating transactions, or describes a multi-step operation where partial completion would corrupt state.

## The problem

Neither `MonadError` nor exceptions reach back and undo side effects that already happened. `ExceptT` short-circuits *the computation*; it can't reverse an `executeQuery` that already ran. Same for `throwIO`.

```haskell
-- ✗ Inconsistent state on failure
transferCredits from to amount = do
  debitAccount from amount    -- ✓ commits to DB
  creditAccount to amount     -- ✗ throws — debit is NOT undone
  -- Money has disappeared.
```

Four approaches, ordered from "use if you can" to "use when you must."

---

## 1. Database transactions (preferred when applicable)

When all side effects happen in one transactional system, wrap the business action in a DB transaction. The database does the rollback for you.

### As a capability

```haskell
class Monad m => MonadTransaction m where
  withTransaction :: m a -> m a
```

### Production instance

The instance must handle three failure modes: domain errors via `MonadError`, synchronous exceptions, and async exceptions (cancellation). `UnliftIO` provides the building blocks.

```haskell
import UnliftIO (bracket_)
import UnliftIO.Exception (try, throwIO, SomeException)

instance MonadTransaction AppM where
  withTransaction action = do
    conn <- asks envConn

    -- Begin the transaction.
    tryIO (beginTransaction conn)

    -- Run the action, watching for both AppError and exceptions.
    -- We need to roll back on EITHER failure mode.
    result <- (Right <$> action)
              `catchError` (\err -> pure (Left err))

    case result of
      Right a -> do
        tryIO (commit conn)
        pure a
      Left err -> do
        tryIO (rollback conn)
        throwError err
```

This handles `MonadError` failures. For async-exception safety you need `bracket`-style cleanup; in practice use the helper your DB library provides (`withTransaction` from `postgresql-simple`, etc.) and wrap *that* in `tryIO`:

```haskell
instance MonadTransaction AppM where
  withTransaction action = do
    conn <- asks envConn
    env  <- ask
    -- The driver's withTransaction handles rollback on exceptions.
    -- We catch AppError manually inside.
    result <- liftIO $ PG.withTransaction conn $ do
      runApp env action
    liftEither' result
```

The driver's `withTransaction` rolls back on any `IO` exception (including async). We thread `AppM`'s `Either AppError a` result through and rethrow.

### Use in business code

```haskell
transferCredits
  :: (MonadAccountDb m, MonadTransaction m, MonadError AppError m)
  => Int -> Int -> Int -> m ()
transferCredits from to amount = withTransaction $ do
  debitAccount from amount
  creditAccount to amount
```

If `creditAccount` raises a domain error or an infra exception, both operations are rolled back. The business code reads like there's no failure path at all.

### Limits

Works only when:
- All side effects are in **one** transactional system (one DB, not two; not DB + Redis; not DB + HTTP).
- The DB supports the isolation level you need.
- The transaction stays short — long transactions hurt throughput and increase deadlock risk.

For anything beyond a single DB, move to Approach 4 (outbox).

---

## 2. Bracket pattern for resources

For acquire/release pairs (file handles, connections, locks, temp directories), use `bracket`. The cleanup runs on success, error, or async exception.

```haskell
import UnliftIO (bracket)

withTempFile :: (FilePath -> AppM a) -> AppM a
withTempFile use = bracket
  (tryIO createTempFile)        -- acquire
  (tryIO . removeFile)           -- release (always runs)
  use                            -- business code

-- Use:
processUpload :: ByteString -> AppM Result
processUpload bytes = withTempFile $ \path -> do
  tryIO (BS.writeFile path bytes)
  parseAndStore path
```

`bracket` is the right tool when:
- Cleanup is always the **same action** (close, delete, release).
- Cleanup must run **even on cancellation** (async-safe).

It's not enough when compensation depends on *what* was done. For that, see sagas.

---

## 3. Sagas — explicit compensation stack

When effects span multiple systems and rollback is custom per step, accumulate compensations and run them in reverse on failure.

### The combinator

```haskell
import Data.IORef
import UnliftIO (newIORef, readIORef, modifyIORef)

newtype Saga m a = Saga { unSaga :: IORef [m ()] -> m a }

-- A saga step: an action plus its compensation, given the action's result.
step :: MonadIO m => m a -> (a -> m ()) -> Saga m a
step action compensate = Saga $ \ref -> do
  result <- action
  liftIO (modifyIORef ref (compensate result :))
  pure result

-- Run a saga: on any error, execute compensations in reverse order.
executeSaga
  :: (MonadIO m, MonadError e m)
  => Saga m a -> m a
executeSaga (Saga f) = do
  ref <- liftIO (newIORef [])
  f ref `catchError` \err -> do
    comps <- liftIO (readIORef ref)
    -- Compensations are pushed onto the front, so the list is already
    -- in reverse-execution order (newest first).
    runCompensations comps
    throwError err
  where
    -- Run all compensations; if one fails, log and continue with the rest.
    -- Aborting on compensation failure leaves the system more inconsistent,
    -- not less.
    runCompensations [] = pure ()
    runCompensations (c:cs) = do
      c `catchError` \_ -> pure ()  -- log here in real code
      runCompensations cs
```

`Saga` is `Functor`/`Applicative`/`Monad` with a bit more work; this minimal version is enough to show the idea.

### Use in business code

```haskell
transferCredits
  :: (MonadAccountDb m, MonadEvents m, MonadError AppError m, MonadIO m)
  => Int -> Int -> Int -> m ()
transferCredits from to amount = executeSaga $ do
  step (debitAccount from amount)
       (\_ -> creditAccount from amount)        -- undo: re-credit
  step (creditAccount to amount)
       (\_ -> debitAccount to amount)           -- undo: re-debit
  step (publishEvent (TransferCompleted from to amount))
       (\_ -> publishEvent (TransferReversed  from to amount))
```

If `publishEvent` fails, the saga runs the compensations:
1. Debit `to` (undo the credit)
2. Credit `from` (undo the debit)

### Gotchas — sagas are easy to get wrong

**Compensations must be idempotent.** A network glitch may cause the compensation to retry. Re-crediting an already-credited account is a bug. Use idempotency keys, conditional updates, or check-then-act under a transaction.

**Compensations can themselves fail.** What then? Three options, all bad:
- *Abort and alert* — leaves the system inconsistent; someone has to clean up.
- *Retry the compensation* — may loop forever if the underlying issue persists.
- *Continue with remaining compensations* — partial rollback, even more confused state.

The code above continues. In production, also log loudly and alert. There is no clean answer.

**Async exceptions during compensation.** If the process is killed mid-rollback, you're stuck. Mitigate by making each compensation itself an idempotent operation that a recovery process can replay.

**Ordering matters.** If step B depends on step A, compensating A before B finishes can fail. The reverse-order rule usually handles this, but not always — sometimes you need a dependency graph.

**Don't use sagas inside a DB transaction.** A transaction already gives you rollback; layering a saga on top is confused. Pick one.

### When to actually use sagas

Reach for sagas when:
- Effects span **multiple transactional systems** (DB + message queue, two DBs).
- Each effect has a well-defined inverse that's safe to run.
- You've considered Approach 4 (outbox) and have a reason it doesn't fit.

Don't use them for:
- Single-DB operations (use Approach 1).
- Operations whose inverses are not safe or not possible (use Approach 4).
- "Just in case" — sagas add real complexity and a class of subtle bugs.

---

## 4. Outbox + eventual consistency

For effects you genuinely cannot undo (sent emails, charged credit cards, published Kafka messages), don't try. Decouple instead.

### The pattern

1. Do all the local-DB work inside one transaction (Approach 1).
2. Inside that same transaction, insert a row into an `outbox` table describing the side effect to perform.
3. A separate worker reads `outbox`, performs the side effect, and marks the row processed. Retries on failure.

```haskell
data OutboxEntry = OutboxEntry
  { entryId      :: UUID
  , entryAction  :: ByteString  -- serialized action descriptor
  , entryStatus  :: Status
  , entryRetries :: Int
  }

class Monad m => MonadOutbox m where
  enqueueOutbox :: OutboxAction -> m ()

confirmOrder
  :: (MonadOrderDb m, MonadOutbox m, MonadTransaction m, MonadError AppError m)
  => OrderId -> m ()
confirmOrder oid = withTransaction $ do
  order <- findOrder oid >>= maybe (throwError (NotFound 0)) pure
  markOrderConfirmed order
  enqueueOutbox (SendEmail (orderEmail order) "Confirmed" (renderConfirmation order))
  enqueueOutbox (ChargeCard (orderCardToken order) (orderTotal order))
  -- Both DB rows commit atomically with the outbox rows, or neither.
```

The user-facing operation succeeds when the transaction commits. The email and charge happen asynchronously, with retries handled by the worker.

### Why this is the production answer

- **Atomicity that crosses systems**: the DB commits the order-state-change and the outbox-entry together. If the process crashes after commit but before the worker runs, the outbox still has the work.
- **Retry without duplication**: outbox entries have IDs; downstream side effects use them as idempotency keys.
- **Backpressure**: the worker controls the rate of external calls.
- **Observability**: stuck outbox entries are visible in the table.

### The worker

A separate Layer 1 process — typically a polling loop or a CDC subscriber on the outbox table. It reads pending entries, dispatches them to the right capability (`MonadEmail`, `MonadPayment`, etc.), and updates status. Use the same three-layer cake structure; the "business logic" of the worker is "dispatch this descriptor to its handler."

### Caveats

- **At-least-once delivery**: handlers must be idempotent. The worker may run the same entry twice.
- **Out-of-order delivery**: if order matters, encode dependencies in the descriptor or use ordered queues.
- **Backpressure on the outbox table**: archive processed entries; don't let it grow forever.

---

## How this fits the three-layer cake

Each approach is a natural extension of the existing pattern.

| Approach | Layer 1 | Layer 2 | Layer 3 |
|----------|---------|---------|---------|
| **Transactions** | `MonadTransaction` instance wraps driver's transaction primitive | Business code uses `withTransaction action` | No change |
| **Bracket** | Resource helpers (`withTempFile`, `withLock`) using `UnliftIO.bracket` | Business code uses helpers | No change |
| **Sagas** | No new Layer 1; compensations call existing capabilities | Saga combinator + `step` registrations | No change |
| **Outbox** | `MonadOutbox` instance writes to outbox table; worker is its own Layer 1 process | Business code calls `enqueueOutbox` inside `withTransaction` | Outbox action descriptors are domain types |

The pattern composes. A realistic business operation often uses several:

```haskell
processCheckout
  :: ( MonadCartDb m, MonadOrderDb m, MonadOutbox m
     , MonadTransaction m, MonadError AppError m )
  => UserId -> m OrderId
processCheckout uid = withTransaction $ do
  cart  <- loadCart uid >>= maybe (throwError EmptyCart) pure
  total <- liftEither' (computeTotal cart)
  order <- createOrder uid cart total
  clearCart uid
  enqueueOutbox (ChargeCard (userCard uid) total)
  enqueueOutbox (SendEmail (userEmail uid) "Order received" ...)
  pure (orderId order)
```

Transaction wraps everything for atomicity. DB-only changes (`createOrder`, `clearCart`) commit or roll back together. External effects go through the outbox, decoupled and retried. No saga needed — the design eliminated the failure mode.

---

## Decision guide

Use this to pick the approach:

1. **Are all effects in one transactional DB?** → Approach 1 (transactions). Stop here.
2. **Is the cleanup always the same action, regardless of what was done?** → Approach 2 (bracket). Stop here.
3. **Do effects span multiple systems but each has a well-defined, safe inverse?** → Consider Approach 3 (saga). But first check Approach 4.
4. **Do effects include things that can't truly be undone (emails sent, money moved externally, messages published)?** → Approach 4 (outbox). This is the answer for most real distributed systems.

When in doubt, prefer Approach 4 over Approach 3. The outbox pattern has fewer subtle failure modes and aligns with how mature systems actually handle cross-service consistency.

---

## Anti-patterns

```haskell
-- ✗ Catching MonadError to "rollback" without an actual rollback mechanism
transferCredits from to amount =
  (debitAccount from amount >> creditAccount to amount)
    `catchError` \err -> do
      -- This isn't a rollback. The debit already happened.
      logError "failed!"
      throwError err
```

```haskell
-- ✗ Mixing saga and DB transaction for the same operation
transferCredits from to amount = withTransaction $ executeSaga $ do
  step (debitAccount from amount) (\_ -> creditAccount from amount)
  step (creditAccount to amount)  (\_ -> debitAccount to amount)
-- The transaction already rolls back on failure.
-- The saga's compensations will run AFTER rollback — re-applying changes.
-- This produces inverted-state corruption.
```

```haskell
-- ✗ Non-idempotent compensation
step (chargeCard token 100)
     (\_ -> chargeCard token (-100))  -- two retries = -200 net
-- ✓ Use a refund operation with idempotency key
step (chargeCard token 100)
     (\chargeId -> refundCharge chargeId)
```

```haskell
-- ✗ Outbox write OUTSIDE the transaction that creates the entity
createOrder uid items = do
  order <- withTransaction (insertOrder uid items)
  enqueueOutbox (SendEmail ...)
  -- If the process dies between the two lines, the order exists but no email queued.
  -- ✓ enqueueOutbox must be INSIDE the same withTransaction.
```

The principle: rollback semantics must match the actual reversibility of the operations. If you can't truly undo it, don't pretend you can — decouple it instead.
