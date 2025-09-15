📝 Notes for Today
1. JWT Signing Methods

* HMAC (HS256 etc.)
  One secret key.
  Same key signs and verifies tokens.
  Simple → good for single backend.

* RSA (RS256 etc.)
  Two keys: Private + Public.
  Private signs, Public verifies.
  Safer for multi-service (auth + APIs).


2. Access Token

* A JWT given to the user after login.
* Short-lived (e.g. 15 min).
* Used on every API request: Authorization: Bearer <token>.
* If expired → cannot call APIs anymore.


3. Refresh Token

* Long-lived (e.g. 7 days).
* Used only to get new access tokens.
* Sent to /refresh endpoint.
* If refresh token valid → server issues new access + new refresh (rotation).
* Makes user experience smooth (no re-login every 15 min).


4. Private Key

* Secret, stays only on server.
* Used to sign JWTs at login/refresh.
* If stolen → attacker can create fake tokens.
* Must never be shared.

5. Public Key

* Generated from private key.
* Safe to publish (/.well-known/jwks.json).
* Used by middleware or other services to verify JWT signatures.
* Cannot create tokens, only verify them.

6. JWT Flow

* User logs in → server validates credentials.
* Server creates JWT (access + refresh).
* Signs JWT with private key.
* Sends tokens to client.
* Client calls API with access token.
* Middleware verifies with public key.
* If expired → client calls /refresh with refresh token.
* Server checks refresh, issues new tokens (signed again with private key).



✅ Quick memory hooks:

HMAC = one key for both sign + verify.

RSA = private signs, public verifies.

Access token = short ticket for APIs.

Refresh token = longer pass to get new access tokens.

Private key = server-only, sign.

Public key = shared, verify.