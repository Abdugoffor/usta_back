package middleware

import "github.com/julienschmidt/httprouter"

type RoleMiddleware func(httprouter.Handle) httprouter.Handle

func CheckRoleGroup(roles ...string) RoleMiddleware {
	return func(next httprouter.Handle) httprouter.Handle {
		return CheckRole(next, roles...)
	}
}
