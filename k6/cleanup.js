import http from 'k6/http';
import { check } from 'k6';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

// Credentials admin — sesuaikan dengan user admin yang ada di database kamu
const ADMIN_EMAIL = __ENV.ADMIN_EMAIL || 'admin@itemly.com';
const ADMIN_PASSWORD = __ENV.ADMIN_PASSWORD || 'Admin@1234';

export const options = {
  vus: 1,
  iterations: 1,
};

export default function () {
  // Login sebagai admin untuk dapat token
  const loginRes = http.post(
    `${BASE_URL}/api/v1/auth/login`,
    JSON.stringify({ email: ADMIN_EMAIL, password: ADMIN_PASSWORD }),
    { headers: { 'Content-Type': 'application/json' } }
  );

  const ok = check(loginRes, {
    'admin login success': (r) => r.status === 200,
  });

  if (!ok) {
    console.error('Admin login failed. Cleanup harus dilakukan manual via SQL:');
    console.error("DELETE FROM users WHERE email LIKE '%@loadtest.com';");
    return;
  }

  console.log('=== CLEANUP ===');
  console.log('k6 tidak bisa jalankan SQL langsung.');
  console.log('Jalankan query ini di database kamu:');
  console.log('');
  console.log("  DELETE FROM users WHERE email LIKE '%@loadtest.com';");
  console.log('');
  console.log('Atau via psql:');
  console.log("  psql -d itemly -c \"DELETE FROM users WHERE email LIKE '%@loadtest.com';\"");
}
