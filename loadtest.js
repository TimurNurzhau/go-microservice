import http from "k6/http";

export const options = {
  scenarios: {
    users: {
      executor: "constant-arrival-rate",
      rate: 1200,
      timeUnit: "1s",
      duration: "60s",
      preAllocatedVUs: 500,
      maxVUs: 1000,
    },
  },
  thresholds: {
    http_req_duration: ["p(95)<10"],
    http_req_failed: ["rate<0.001"],
    dropped_iterations: ["count==0"],
  },
};

export default function () {
  http.get("http://127.0.0.1:8080/api/users");
}
