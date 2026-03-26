/*
 * Copyright 2025 coze-dev Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/coze-dev/coze-studio/backend/api/internal/httputil"
	"github.com/coze-dev/coze-studio/backend/application/user"
	"github.com/coze-dev/coze-studio/backend/bizpkg/config"
	"github.com/coze-dev/coze-studio/backend/domain/user/entity"
	"github.com/coze-dev/coze-studio/backend/pkg/ctxcache"
	"github.com/coze-dev/coze-studio/backend/pkg/errorx"
	"github.com/coze-dev/coze-studio/backend/pkg/logs"
	"github.com/coze-dev/coze-studio/backend/types/consts"
	"github.com/coze-dev/coze-studio/backend/types/errno"
)

var noNeedSessionCheckPath = map[string]bool{
	"/api/passport/web/email/login/":       true,
	"/api/passport/web/email/register/v2/": true,
	"/api/auth/third_login":                true,
	"/api/open/connect/session":            true,
}

func SessionAuthMW() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		path := string(ctx.GetRequest().URI().Path())
		method := string(ctx.GetRequest().Method())

		fmt.Printf("[SessionAuthMW] Request path: %s, method: %s\n", path, method)

		if (path == "/api/auth/third_login" || path == "/api/auth/third_login/") && method == "GET" {
			fmt.Printf("[SessionAuthMW] Third-party login path detected: %s, method: %s\n", path, method)
			ctx.Next(c)
			return
		}

		if (path == "/api/open/connect/session" || path == "/api/open/connect/session/") && method == "GET" {
			fmt.Printf("[SessionAuthMW] Connect session path detected: %s, method: %s\n", path, method)
			ctx.Next(c)
			return
		}

		if noNeedSessionCheckPath[path] {
			fmt.Printf("[SessionAuthMW] No session check path detected: %s\n", path)
			ctx.Next(c)
			return
		}

		requestAuthType := ctx.GetInt32(RequestAuthTypeStr)
		fmt.Printf("[SessionAuthMW] Request auth type: %d\n", requestAuthType)
		if requestAuthType != int32(RequestAuthTypeWebAPI) {
			fmt.Printf("[SessionAuthMW] Not web API request, skipping session check\n")
			ctx.Next(c)
			return
		}

		s := ctx.Cookie(entity.SessionKey)
		fmt.Printf("[SessionAuthMW] Session key: %s\n", s)
		if len(s) == 0 {
			fmt.Printf("[SessionAuthMW] Session key is empty, returning 401\n")
			httputil.Unauthorized(ctx, "missing session_key in cookie")
			return
		}

		session, err := user.UserApplicationSVC.ValidateSession(c, string(s))
		if err != nil {
			fmt.Printf("[SessionAuthMW] Validate session failed: %v\n", err)
			httputil.InternalError(c, ctx, err)
			return
		}

		if session != nil {
			fmt.Printf("[SessionAuthMW] Session validated successfully: %v\n", session)
			ctxcache.Store(c, consts.SessionDataKeyInCtx, session)
		} else {
			fmt.Printf("[SessionAuthMW] Session is nil\n")
		}

		ctx.Next(c)
	}
}

func AdminAuthMW() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		session, ok := ctxcache.Get[*entity.Session](c, consts.SessionDataKeyInCtx)
		if !ok {
			logs.Errorf("[AdminAuthMW] session data is nil")
			httputil.InternalError(c, ctx,
				errorx.New(errno.ErrUserAuthenticationFailed, errorx.KV("reason", "session data is nil")))
			return
		}

		baseConf, err := config.Base().GetBaseConfig(c)
		if err != nil {
			logs.Errorf("[AdminAuthMW] get base config failed, err: %v", err)
			httputil.InternalError(c, ctx, err)
			return
		}

		if baseConf.AdminEmails == "" {
			logs.CtxWarnf(c, "[AdminAuthMW] admin emails is empty")
			ctx.Next(c)
			return
		}

		adminEmails := strings.Split(baseConf.AdminEmails, ",")
		for _, adminEmail := range adminEmails {
			if strings.EqualFold(adminEmail, session.UserEmail) {
				ctx.Next(c)
				return
			}
		}

		httputil.Unauthorized(ctx, "the account does not have permission to access")
	}
}
