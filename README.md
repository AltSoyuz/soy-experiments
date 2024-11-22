# GOKISS

GOKISS is a simple and easy-to-use web server written in Go. It is designed to serve static files or create simple web applications effortlessly.

## Installation

To install GOKISS, ensure you have Go installed on your system. You can download Go from the official website [here](https://golang.org/). Version 1.22 or higher is required.

## Resources


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

## User Scenarios

1. New User Registration: User creates an account and initiates email verification.

2. Login with Username and Password: User logs in, and a session is created.

3. Email Verification Process: Ensures the user has verified their email before gaining full access.

4. Password Reset: Users can request to reset their password if they forget it.

5. OAuth Login: User logs in using third-party authentication, linking their account.

6. Logout: Users can end their session explicitly.
Session Expiry Management: Background cleanup of expired sessions.



1. User Registration
Steps:

User submits username, password, and other required information.
Server:
Hashes the password using a secure hashing algorithm (e.g., bcrypt).
Inserts a record into the user table with email_verified = 0.
Outcome:

A new user record is created.
An email verification request is initiated (see scenario 3).

2. User Login (Password-Based)
Steps:

User submits username and password.
Server:
Fetches the password_hash for the given username from the user table.
Verifies the submitted password against the password_hash.
If successful:
Creates a new record in the session table with an expiry timestamp (expires_at).
Outcome:

User is logged in with a valid session.

3. Email Verification
Steps:

After registration, the system generates:
A unique code for email verification.
created_at and expires_at timestamps.
Inserts a record into the email_verification_request table.
Sends an email to the user with the verification code.
User clicks the link or enters the code.
Server:
Validates the code and checks the expires_at timestamp.
Updates email_verified to 1 in the user table upon success.
Outcome:

User email is verified.

4. Password Reset Request
Steps:

User initiates a password reset request by providing their registered email.
Server:
Generates a unique code_hash (hashed version of the reset code).
Calculates expires_at timestamp.
Inserts a record into the password_reset_request table.
Sends the reset code to the user's email.
User submits the reset code and new password.
Server:
Verifies the reset code by matching its hash with code_hash.
Ensures the expires_at timestamp is valid.
Updates the password_hash in the user table with the new hashed password.
Outcome:

User password is reset.

5. OAuth Login (Third-Party Authentication)
Steps:

User initiates login via an OAuth provider (e.g., Google, Facebook).
Server:
Receives provider, provider_user_id, and other user data from the OAuth provider.
Checks if the provider_user_id exists in the oauth_accounts table.
If not:
Creates a record in the oauth_accounts table.
Links the OAuth account to a user record.
If it exists:
Fetches the associated user_id and creates a new session in the session table.
Outcome:

User is logged in using their OAuth account.

6. Logout
Steps:

User logs out by invalidating their session.
Server deletes the session record from the session table for the given id.
Outcome:

User session is invalidated.

7. Session Expiry
Steps:

Periodic cleanup or upon user request:
Server checks all records in the session table where expires_at is in the past.
Deletes expired sessions.
Outcome:

Expired sessions are purged.
