// Code generated by microgen 0.9.0. DO NOT EDIT.

package service

import (
	"context"
	"fmt"
	log "github.com/go-kit/kit/log"
	service "github.com/recolabs/microgen/examples/usersvc/pkg/usersvc"
)

// RecoveringMiddleware recovers panics from method calls, writes to provided logger and returns the error of panic as method error.
func RecoveringMiddleware(logger log.Logger) Middleware {
	return func(next service.UserService) service.UserService {
		return &recoveringMiddleware{
			logger: logger,
			next:   next,
		}
	}
}

type recoveringMiddleware struct {
	logger log.Logger
	next   service.UserService
}

func (M recoveringMiddleware) CreateUser(ctx context.Context, user service.User) (id string, err error) {
	defer func() {
		if r := recover(); r != nil {
			M.logger.Log("method", "CreateUser", "message", r)
			err = fmt.Errorf("%v", r)
		}
	}()
	return M.next.CreateUser(ctx, user)
}

func (M recoveringMiddleware) UpdateUser(ctx context.Context, user service.User) (err error) {
	defer func() {
		if r := recover(); r != nil {
			M.logger.Log("method", "UpdateUser", "message", r)
			err = fmt.Errorf("%v", r)
		}
	}()
	return M.next.UpdateUser(ctx, user)
}

func (M recoveringMiddleware) GetUser(ctx context.Context, id string) (user service.User, err error) {
	defer func() {
		if r := recover(); r != nil {
			M.logger.Log("method", "GetUser", "message", r)
			err = fmt.Errorf("%v", r)
		}
	}()
	return M.next.GetUser(ctx, id)
}

func (M recoveringMiddleware) FindUsers(ctx context.Context) (results map[string]service.User, err error) {
	defer func() {
		if r := recover(); r != nil {
			M.logger.Log("method", "FindUsers", "message", r)
			err = fmt.Errorf("%v", r)
		}
	}()
	return M.next.FindUsers(ctx)
}

func (M recoveringMiddleware) CreateComment(ctx context.Context, comment service.Comment) (id string, err error) {
	defer func() {
		if r := recover(); r != nil {
			M.logger.Log("method", "CreateComment", "message", r)
			err = fmt.Errorf("%v", r)
		}
	}()
	return M.next.CreateComment(ctx, comment)
}

func (M recoveringMiddleware) GetComment(ctx context.Context, id string) (comment service.Comment, err error) {
	defer func() {
		if r := recover(); r != nil {
			M.logger.Log("method", "GetComment", "message", r)
			err = fmt.Errorf("%v", r)
		}
	}()
	return M.next.GetComment(ctx, id)
}

func (M recoveringMiddleware) GetUserComments(ctx context.Context, userId string) (list []service.Comment, err error) {
	defer func() {
		if r := recover(); r != nil {
			M.logger.Log("method", "GetUserComments", "message", r)
			err = fmt.Errorf("%v", r)
		}
	}()
	return M.next.GetUserComments(ctx, userId)
}
