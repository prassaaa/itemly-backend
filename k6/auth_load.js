import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('errors');

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export const options = {
  stages: [
    { duration: '10s', target: 3 },
    { duration: '30s', target: 3 },
    { duration: '5s', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<1000'],
    errors: ['rate<0.3'],
  },
};

function retryPost(url, body, params, maxRetries) {
  for (let i = 0; i < maxRetries; i++) {
    const res = http.post(url, body, params);
    if (res.status !== 429) {
      return res;
    }
    sleep(1 + i);
  }
  return http.post(url, body, params);
}

export default function () {
  const uniqueId = `${__VU}-${__ITER}-${Date.now()}`;
  const email = `user_${uniqueId}@loadtest.com`;
  const username = `user_${uniqueId}`;
  const password = 'Test@1234';
  const jsonHeaders = { headers: { 'Content-Type': 'application/json' } };

  // Register
  const registerRes = retryPost(
    `${BASE_URL}/api/v1/auth/register`,
    JSON.stringify({ username, email, password }),
    jsonHeaders,
    3
  );
  const registerOk = check(registerRes, {
    'register status is 201': (r) => r.status === 201,
  });
  errorRate.add(!registerOk);
  if (!registerOk) {
    sleep(2);
    return;
  }

  const registerBody = registerRes.json();
  const accessToken = registerBody.access_token;
  const refreshToken = registerBody.refresh_token;

  sleep(1.5);

  // Login
  const loginRes = retryPost(
    `${BASE_URL}/api/v1/auth/login`,
    JSON.stringify({ email, password }),
    jsonHeaders,
    3
  );
  const loginOk = check(loginRes, {
    'login status is 200': (r) => r.status === 200,
  });
  errorRate.add(!loginOk);

  const loginToken = loginOk ? loginRes.json().access_token : accessToken;

  sleep(1.5);

  // Get Profile (no rate limit on this endpoint)
  const profileRes = http.get(`${BASE_URL}/api/v1/profile`, {
    headers: { Authorization: `Bearer ${loginToken}` },
  });
  check(profileRes, {
    'profile status is 200': (r) => r.status === 200,
  });
  errorRate.add(profileRes.status !== 200);

  sleep(1.5);

  // Refresh Token
  const refreshRes = retryPost(
    `${BASE_URL}/api/v1/auth/refresh`,
    JSON.stringify({ refresh_token: refreshToken }),
    jsonHeaders,
    3
  );
  const refreshOk = check(refreshRes, {
    'refresh status is 200': (r) => r.status === 200,
  });
  errorRate.add(!refreshOk);

  const newAccessToken = refreshOk
    ? refreshRes.json().access_token
    : loginToken;

  sleep(1.5);

  // Logout (no rate limit on this endpoint)
  const logoutRes = http.post(`${BASE_URL}/api/v1/auth/logout`, null, {
    headers: { Authorization: `Bearer ${newAccessToken}` },
  });
  check(logoutRes, {
    'logout status is 200': (r) => r.status === 200,
  });
  errorRate.add(logoutRes.status !== 200);

  sleep(2);
}
