package api

import (
	"database/sql"
	db "input-backend-master-class/db/generated"
	"input-backend-master-class/util"
	"input-backend-master-class/worker"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
)

type createUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
}

type userResponse struct {
	Username         string `json:"username"`
	FullName         string `json:"full_name"`
	Email            string `json:"email"`
	PasswordChangeAt string `json:"password_change_at"`
	CreatedAt        string `json:"created_at"`
}

func newUserResponse(user db.user) userResponse {
	return userResponse{
		Username:         user.Username,
		FullName:         user.FullName,
		Email:            user.Email,
		PasswordChangeAt: user.PasswordChangeAt.Format(time.RFC3339),
		CreatedAt:        user.CreatedAt.Format(time.RFC3339),
	}
}

type loginUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type loginUserResponse struct {
	AccessToken string `json:"access_token"`
	userResponse
}

func (server *Server) createUser(ctx *gin.Context) {
	var req createUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	arg := db.CreateUserTxParams{
		CreateUserParams: db.CreateUserParams{
			Username:       req.Username,
			HashedPassword: hashedPassword,
			FullName:       req.FullName,
			Email:          req.Email,
		},
		AfterCreate: func(user db.User) error {
			taskPayload := &worker.PayloadSendVerifyEmail{
				UserName: user.Username,
			}
			opts := []asynq.Option{
				asynq.MaxRetry(10),
				asynq.Timeout(10 * time.Second),
				asynq.Queue(worker.QueueCritical),
			}
			return server.taskDistributor.DistributeTaskSendVerifyEmail(ctx, taskPayload, opts...)
		},
	}

	// user, err := server.store.CreateUser(ctx, arg)
	txResult, err := server.store.CreateUserTx(ctx, arg)
	if err != nil {
		if db.ErrorCode(err) == db.UniqueViolation {
			ctx.JSON(http.StatusForbidden, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// ここでtaskを作成する
	// userの作成とtaskはトランザクションで対応する必要がある。 失敗したらrollbackする
	// taskPayload := &worker.PayloadSendVerifyEmail{
	// 	UserName: user.Username,
	// }
	// opts := []asynq.Option{
	// 	asynq.MaxRetry(10),
	// 	asynq.Timeout(10 * time.Second),
	// 	asynq.Queue(worker.QueueCritical),
	// }
	// err = server.taskDistributor.DistributeTaskSendVerifyEmail(ctx, taskPayload, opts...)
	// if err != nil {
	// 	// return nil, status.Errorf(codes.Internal, "failed to distribute task: %v", err)
	// 	ctx.JSON(http.StatusInternalServerError, errorResponse(err))
	// 	return
	// }

	rsp := newUserResponse(txResult.User)
	ctx.JSON(http.StatusOK, rsp)
}

func (server *Server) loginUser(ctx *gin.Context) {
	var req loginUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user, err := server.store.GetUser(req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	err = util.CheckPassword(req.Password, user.PasswordHash)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	accessToken, err := server.tokenMaker.CreateToken(
		user.Username,
		server.config.AccessTokenDuration,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := loginUserResponse{
		AccessToken:  accessToken,
		userResponse: newUserResponse(user),
	}
	ctx.JSON(http.StatusOK, rsp)

}
