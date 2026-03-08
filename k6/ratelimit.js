import http from 'k6/http';
import { check } from 'k6';
import { Counter } from 'k6/metrics';

const rateLimited = new Counter('rate_limited_responses');

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export const options = {
  vus: 1,
  iterations: 20,
  thresholds: {
    rate_limited_responses: ['count>=1'],
  },
};

export default function () {
  const res = http.post(
    `${BASE_URL}/api/v1/auth/login`,
    JSON.stringify({
      email: 'ratelimit@test.com',
      password: 'Test@1234',
    }),
    { headers: { 'Content-Type': 'application/json' } }
  );

  if (res.status === 429) {
    rateLimited.add(1);
  }

  check(res, {
    'status is 401 or 429': (r) => r.status === 401 || r.status === 429,
  });
}
