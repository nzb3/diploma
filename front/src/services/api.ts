import axios from 'axios';
import {AskRequest, AskResponse, Resource, SaveDocumentRequest} from '../types/api';
import authService from './authService';

const API_BASE_URL = 'http://search:8080/api/v1';


const api = axios.create({
  baseURL: API_BASE_URL,
  withCredentials: true,
  headers: {
    'Content-Type': 'application/json',
  }
});

api.interceptors.request.use(
  async (config) => {
    if (authService.isAuthenticated()) {
      const token = authService.getToken();
      if (token) {
        config.headers.Authorization = `Bearer ${token}`;
      }
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// Add interceptor to handle 401/403 responses
api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;
    
    if ((error.response?.status === 401 || error.response?.status === 403) &&
        !originalRequest._retry && 
        authService.isAuthenticated()) {
      
      originalRequest._retry = true;
      
      try {
        const refreshed = await authService.updateToken(10);
        
        if (refreshed) {
          const token = authService.getToken();
          if (token) {
            originalRequest.headers.Authorization = `Bearer ${token}`;
          }
          return axios(originalRequest);
        }
      } catch (refreshError) {
        console.error('Failed to refresh token', refreshError);
        authService.login();
      }
    }
    
    return Promise.reject(error);
  }
);

// Helper function to create an EventSource that works with interceptors
const createSSEConnection = async (url: string): Promise<EventSource> => {
  // First, apply the same authentication logic from interceptors
  const fullUrl = `${window.location.origin}${API_BASE_URL}${url}`;
  const urlObj = new URL(fullUrl);
  
  if (authService.isAuthenticated()) {
    const token = authService.getToken();
    if (token) {
      // Add token as URL parameter so EventSource can use it
      urlObj.searchParams.append('auth_token', token);
    }
  }
  
  // Create a native EventSource with the authenticated URL
  return new EventSource(urlObj.toString(), {
    withCredentials: true
  });
};

export const getResources = async (): Promise<Resource[]> => {
  const response = await api.get('/resources');
  return response.data;
};

export const getResource = async (id: string): Promise<Resource> => {
  const response = await api.get(`/resources/${id}`);
  return response.data;
};

export const deleteResource = async (id: string): Promise<void> => {
  await api.delete(`/resources/${id}`);
};

export const saveResource = async (data: SaveDocumentRequest): Promise<EventSource> => {
  // First, upload the resource
  await api.post('/resources', data);
  
  // Then create SSE connection to track processing
  return createSSEConnection('/resources');
};

export const askQuestion = async (data: AskRequest): Promise<AskResponse> => {
  const response = await api.post('/ask', data);
  return response.data;
};

export const streamAnswer = async (question: string): Promise<EventSource> => {
  // Create a combined endpoint that handles both posting the question and streaming the answer
  const queryParams = new URLSearchParams({ question });
  
  return createSSEConnection(`/ask/stream?${queryParams.toString()}`);
};

export const cancelStream = async (processId: string): Promise<void> => {
  await api.delete(`/ask/stream/cancel/${processId}`);
}; 