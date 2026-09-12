import http from 'k6/http'; import { check } from 'k6';
export const options = { vus: 20, duration: '30s', thresholds: { http_req_duration: ['p(95)<150'] } };
export default function() {
  const key = `k6-${__VU}-${__ITER}`;
  const res = http.post('http://localhost:8080/api/v1/orders', JSON.stringify({items:[{product_id:'PROD-001',qty:1}]}), {headers:{'Content-Type':'application/json','Idempotency-Key':key}});
  check(res, {'201 or 200': r => r.status===201||r.status===200});
}
