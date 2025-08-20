import axios from 'axios';

const API_BASE_URL = 'http://localhost:8080/api';

const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Add auth token to requests
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Handle auth errors
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Clear auth data and redirect to login
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

// Swimmers API
export const swimmersAPI = {
  getAll: () => api.get('/swimmers'),
  getById: (id) => api.get(`/swimmers/${id}`),
  create: (swimmer) => api.post('/swimmers', swimmer),
};

// Times API
export const timesAPI = {
  getAll: () => api.get('/times'),
  getBySwimmer: (swimmerId) => api.get(`/times/${swimmerId}`),
  create: (time) => api.post('/times', time),
};

// Events API
export const eventsAPI = {
  getAll: () => api.get('/events'),
};

// Strokes API
export const strokesAPI = {
  getAll: () => api.get('/strokes'),
};

export default api;
