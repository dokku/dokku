# Routing coverage

Eight browser cases exercise nginx routing to non-web processes.

| Case | Behavior checked |
| --- | --- |
| TC-01 | A browser fetch renders the API's configured identity marker. |
| TC-02 | After a process restart, the marker remains and the previous container hostname disappears. |
| TC-03 | After configuration rollout, the API renders the new marker rather than the previous value. |
| TC-04 | Removing the route makes the browser fetch show the web fallback's HTTP 404. |
| TC-05 | Re-adding the route restores the API response. |
| TC-06 | A live WebSocket echoes a generated message, the backend marker and the expected Host header. |
| TC-07 | Scaling API to zero renders Dokku's bundled error page, with HTTP 502 confirmed independently. |
| TC-08 | With prefix stripping enabled, the browser follows the API's prefix-aware redirect to its login page. |

## Why browser coverage

Native tests in `tests/unit/proxy-route-nginx.bats` cover routing behavior and
rendered configuration. Basic routing and removal overlap the browser suite's
setup checks. This suite additionally executes fixture JavaScript, opens a live
WebSocket and follows redirects through a real nginx deployment. Lifecycle cases
check the browser-visible result after restart, configuration and route changes.

The fixture sets `E2E_EXTERNAL_PREFIX=/api/v0` so the API constructs
`Location: /api/v0/login`. TC-08 tests navigation with an application that knows
its public prefix; it does not require proxy redirect rewriting.

## Assertions and cleanup

Readiness uses web `/healthz` and `nginx -t`, independently of case outcomes.
Expected markers come from configuration, the prior hostname from Docker
inspection and error-page text from the repository template.
HTTP and WebSocket checks run after each browser verdict; disagreement fails
the controller. Cleanup restores and verifies the marker, process scale and
original routes, including when a browser assertion fails.

Manual wrong-state controls omit the relevant change or alter its route.
They must reach a failing browser assertion and agree with the independent
check; a setup error is not a successful negative control.
The redirect control serves an HTTP 200 page with the wrong heading to check
the login assertion rather than an unrelated navigation failure.

Coverage is limited to nginx, HTTP and the selected desktop browser profile.
It does not replace the native suite or establish TLS, other proxy support
or continuous availability during a process restart.
