package middleware

import "github.com/questx-lab/backend/pkg/xcontext"

func AllowCors(ctx xcontext.Context) {
	if origin := ctx.Request().Header.Get("Origin"); origin != "" {
		ctx.Writer().Header().Set("Access-Control-Allow-Origin", "*")
		ctx.Writer().Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		ctx.Writer().Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	}
}
