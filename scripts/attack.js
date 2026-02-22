import http from 'k6/http';
import { check } from 'k6';

export let options = {
  vus: 200, 
  duration: '5s', 
};

export default function () {
  const url = "http://localhost:8080/purchase-async"
  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  let res = http.post(url, JSON.stringify({}), params);
  check(res, { 'status was 200': (r) => r.status == 200 });
}