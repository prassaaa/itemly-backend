import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('errors');

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export const options = {
  stages: [
    { duration: '30s', target: 10 },
    { duration: '1m', target: 10 },
    { duration: '10s', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'],
    errors: ['rate<0.1'],
  },
};

export default function () {
  const uniqueId = `${__VU}-${__ITER}-${Date.now()}`;
  const email = `user_${uniqueId}@loadtest.com`;
  const username = `user_${uniqueId}`;
  const password = 'Test@1234';

  // Register
  const registerRes = http.post(
    `${BASE_URL}/api/v1/auth/register`,
    JSON.stringify({ username, email, password }),
    { headers: { 'Content-Type': 'application/json' } }
  );
  const registerOk = check(registerRes, {
    'register status is 201': (r) => r.status === 201,
  });
  errorRate.add(!registerOk);
  if (!registerOk) {
    return;
  }

  const registerBody = registerRes.json();
  const accessToken = registerBody.access_token;
  const refreshToken = registerBody.refresh_token;

  sleep(0.5);

  // Login
  const loginRes = http.post(
    `${BASE_URL}/api/v1/auth/login`,
    JSON.stringify({ email, password }),
    { headers: { 'Content-Type': 'application/json' } }
  );
  check(loginRes, {
    'login status is 200': (r) => r.status === 200,
  });
  errorRate.add(loginRes.status !== 200);

  const loginToken = loginRes.json().access_token || accessToken;

  sleep(0.5);

  // Get Profile
  const profileRes = http.get(`${BASE_URL}/api/v1/profile`, {
    headers: { Authorization: `Bearer ${loginToken}` },
  });
  check(profileRes, {
    'profile status is 200': (r) => r.status === 200,
  });
  errorRate.add(profileRes.status !== 200);

  sleep(0.5);

  // Refresh Token
  const refreshRes = http.post(
    `${BASE_URL}/api/v1/auth/refresh`,
    JSON.stringify({ refresh_token: refreshToken }),
    { headers: { 'Content-Type': 'application/json' } }
  );
  check(refreshRes, {
    'refresh status is 200': (r) => r.status === 200,
  });
  errorRate.add(refreshRes.status !== 200);

  const newAccessToken =
    refreshRes.status === 200
      ? refreshRes.json().access_token
      : loginToken;

  sleep(0.5);

  // Logout
  const logoutRes = http.post(`${BASE_URL}/api/v1/auth/logout`, null, {
    headers: { Authorization: `Bearer ${newAccessToken}` },
  });
  check(logoutRes, {
    'logout status is 200': (r) => r.status === 200,
  });
  errorRate.add(logoutRes.status !== 200);

  sleep(1);
}
