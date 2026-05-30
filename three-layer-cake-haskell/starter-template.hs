-- Three-Layer Cake starter template
--
-- Copy this file and adapt:
--   1. Replace Todo with your domain type(s) in Layer 3.
--   2. Replace AppError constructors with your domain errors.
--   3. Replace Conn with your database driver's connection type.
--   4. Replace the *Sql stubs with calls to your driver (postgresql-simple, etc.).
--   5. Add capability classes for any additional effects (MonadEmail, MonadHttp...).
--
-- See SKILL.md for the full pattern, testing.md for tests,
-- alternative-exceptions.md for the exception-based variant,
-- handles-upgrade.md for adding handles to typeclass instances.

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

-- ════════════════════════════════════════════════════════════
-- LAYER 3: Pure core
-- ════════════════════════════════════════════════════════════

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
  -- Add domain-specific error constructors here.
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

-- ════════════════════════════════════════════════════════════
-- LAYER 1: Capability typeclasses
-- ════════════════════════════════════════════════════════════

class Monad m => MonadTodoDb m where
  findTodo   :: Int  -> m (Maybe Todo)
  insertTodo :: Text -> m Todo
  updateTodo :: Todo -> m ()

class Monad m => MonadLogger m where
  logInfo  :: Text -> m ()
  logError :: Text -> m ()

-- Add more capability classes here as you grow the app.

-- ════════════════════════════════════════════════════════════
-- LAYER 2: The app monad
-- ════════════════════════════════════════════════════════════

-- Replace `Conn` with your driver's actual connection type
-- (e.g. Database.PostgreSQL.Simple.Connection).
data Conn = Conn

data Env = Env
  { envConn :: Conn
  -- Add more resources as needed (HTTP managers, configs, refs).
  }

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

-- ════════════════════════════════════════════════════════════
-- The IO boundary
-- ════════════════════════════════════════════════════════════

-- `try` from UnliftIO.Exception only catches synchronous exceptions —
-- async exceptions (ThreadKilled, UserInterrupt, timeouts) still propagate.
tryIO :: IO a -> AppM a
tryIO action = liftIO (try action) >>= \case
  Right a                   -> pure a
  Left (e :: SomeException) -> throwError (classify e)
  where
    -- Refine this to map specific exception types to specific AppError variants
    -- (e.g. SqlError → DbError with code, IOException → DbError, etc.).
    classify e = DbError (T.pack (show e))

-- One withX helper per resource. Pulls the resource from Env, runs the
-- IO action, lifts exceptions into AppError.
withDb :: (Conn -> IO a) -> AppM a
withDb f = asks envConn >>= tryIO . f

-- ════════════════════════════════════════════════════════════
-- LAYER 1: Production instances
-- ════════════════════════════════════════════════════════════

-- Replace these stubs with calls into your DB library:
--   findTodoSql   :: Conn -> Int  -> IO (Maybe Todo)
--   insertTodoSql :: Conn -> Text -> IO Todo
--   updateTodoSql :: Conn -> Todo -> IO ()

findTodoSql   :: Conn -> Int  -> IO (Maybe Todo)
findTodoSql   _ _ = pure Nothing

insertTodoSql :: Conn -> Text -> IO Todo
insertTodoSql _ t = pure (Todo 1 t False)

updateTodoSql :: Conn -> Todo -> IO ()
updateTodoSql _ _ = pure ()

instance MonadTodoDb AppM where
  findTodo   tid   = withDb $ \c -> findTodoSql   c tid
  insertTodo title = withDb $ \c -> insertTodoSql c title
  updateTodo todo  = withDb $ \c -> updateTodoSql c todo

instance MonadLogger AppM where
  logInfo  msg = liftIO (putStrLn ("[INFO]  " ++ T.unpack msg))
  logError msg = liftIO (putStrLn ("[ERROR] " ++ T.unpack msg))

-- ════════════════════════════════════════════════════════════
-- LAYER 2: Business logic
-- ════════════════════════════════════════════════════════════

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

-- ════════════════════════════════════════════════════════════
-- The HTTP boundary (framework-agnostic shape)
-- ════════════════════════════════════════════════════════════

-- Run a business action and map AppError to an HTTP-style response.
-- Adapt to your framework's Response/Status types.

-- Stubs for the response shape:
data Response = Response { status :: Int, body :: Text } deriving Show

errorToResponse :: AppError -> IO Response
errorToResponse err = case err of
  InvalidTitle msg     -> pure (Response 400 msg)
  NotFound tid         -> pure (Response 404 (T.pack (show tid)))
  AlreadyCompleted tid -> pure (Response 409 (T.pack (show tid)))
  DbError msg -> do
    -- Log infra details internally; do NOT leak to the client.
    putStrLn ("[ERROR] " ++ T.unpack msg)
    pure (Response 500 "internal error")

runHandler :: (a -> IO Response) -> Env -> AppM a -> IO Response
runHandler onSuccess env action = do
  result <- runApp env action
  either errorToResponse onSuccess result

-- Example handlers:
handleCreate :: Env -> Text -> IO Response
handleCreate env title =
  runHandler (\todo -> pure (Response 201 (todoTitle todo)))
             env
             (createTodo title)

handleComplete :: Env -> Int -> IO Response
handleComplete env tid =
  runHandler (\todo -> pure (Response 200 (todoTitle todo)))
             env
             (completeTodo tid)
