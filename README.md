# SOY-EXPERIMENTS 

## Description

This repo has contains a collection of experiments principally in Go. The goal here is to learn and experiment with different technologies, patterns, and best practices. The experiments are not meant to be production-ready but rather a learning experience.
Some links to resources are provided at the end of this document.

### [List of Experiments](#experiments) 
- [x] Session auth
- [x] Email verification with TOTP
- [x] Server-side rendering
- [x] Rate limiting
- [x] Sqlite3
- [x] Db migrations
- [x] Logging
- [x] Middleware
- [x] Configuration
- [x] Unit Testing
- [x] Integration Testing
- [x] Automatisaton with makefiles
- [x] CSRF protection (origin check but not token in form) 
- [] OAuth login
- [] Password reset
- [] Grpc with protobuf
- [] ConnectRPC
- [] React frontend
- [] TailwindCSS
- [] Pulumi deployment on Hetzner
- [] Github actions
- [] Realtime with websockets
- [] Realtime with SSE
- [] Realtime with GRPC
- [] E2E testing with Cypress/Playwright
- [] SSR with React
- [] [Request Coallescing](https://jazco.dev/2023/09/28/request-coalescing/)
- [] Grafana & Prometheus
- [] Memcached
- [] Kubernetes
- [] Temporal io
- [] Payment
- [] Live reload (Optional)

## Installation

Ensure you have Go installed on your system. You can download Go from the official website [here](https://golang.org/). Version 1.22 or higher is required.

## Some Resources

**Structure & Language:**
- [Structure](https://go.dev/doc/modules/layout)
- [Writing Web Applications](https://golang.org/doc/articles/wiki/)
- [Guidelines](https://google.github.io/styleguide/go/best-practices)

**Logging:**
- [Official](https://pkg.go.dev/log)
- [Slog Guide](https://betterstack.com/community/guides/logging/logging-in-go/#getting-started-with-slog)

**Http:**
- [HTTP Services](https://grafana.com/blog/2024/02/09/how-i-write-http-services-in-go-after-13-years/#maker-funcs-return-the-handler)

**Testing:**
- [Test without mocks](https://quii.gitbook.io/learn-go-with-tests/testing-fundamentals/working-without-mocks)
- [Test with functions](https://itnext.io/f-tests-as-a-replacement-for-table-driven-tests-in-go-8814a8b19e9e)
- [Test with tables](https://go.dev/wiki/TableDrivenTests)

**Auth:**
- [Faore](https://faroe.dev/) 
- [What kind of auth?](https://pilcrowonpaper.com/blog/how-i-would-do-auth/)
- [Lucia](https://lucia-auth.com/sessions/basic-api/sqlite)
- [Auth middleware](https://pilcrowonpaper.com/blog/middleware-auth/)
- [Rate limit](https://go.dev/wiki/RateLimiting)

## Exemples
- [Neosync](https://github.com/nucleuscloud/neosync)
- [Kutt](https://github.com/thedevs-network/kutt)
- [VictoriaMetrics](https://github.com/VictoriaMetrics/VictoriaMetrics)

## Inspiration
- [Company process I like](https://betterstack.com/careers)

## UI
- [UI framework comparaison](https://component-party.dev/)
