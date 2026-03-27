import http from 'k6/http';
import { check } from 'k6';

export const options = {
  scenarios: {
    constant_request_rate: {
      executor: 'constant-arrival-rate',
    
      rate: 10000, 
      timeUnit: '1s', 
      duration: '30s',
      preAllocatedVUs: 1000, 
      maxVUs: 5000, 
    },
  },
};

export default function () {
  const url = 'http://api:8080/purchase-async';
  
  const params = {
    headers: { 'Content-Type': 'application/json' },
  };

  http.post(url, JSON.stringify({}), params);
}