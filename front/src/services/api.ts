import axios from 'axios';
import { Resource, AskRequest, AskResponse, SaveDocumentRequest } from '../types/api';

// In production (Docker), use relative URLs for API requests
// In development, use the full URL to the backend
const isDevEnvironment = window.location.hostname === 'localhost' && window.location.port !== '80';
const API_BASE_URL = isDevEnvironment 
  ? 'http://search.deltanotes.orb.local:8080/api/v1'
  : '/api/v1';

const api = axios.create({
  baseURL: API_BASE_URL,
  withCredentials: true,
  headers: {
    'Content-Type': 'application/json',
  }
});

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
  await api.post('/resources', data);
  
  // Use the same base URL logic for EventSource
  const eventSourceUrl = isDevEnvironment
    ? `${API_BASE_URL}/resources`
    : `${window.location.origin}${API_BASE_URL}/resources`;
    
  const eventSource = new EventSource(eventSourceUrl, {
    withCredentials: true,
  });
  
  return eventSource;
};

export const askQuestion = async (data: AskRequest): Promise<AskResponse> => {
  const response = await api.post('/ask', data);
  return response.data;
};

export const streamAnswer = (question: string): EventSource => {
  const url = new URL(
      isDevEnvironment
          ? `${API_BASE_URL}/ask/stream`
          : `${window.location.origin}${API_BASE_URL}/ask/stream`
  );

  url.searchParams.append('question', question);;

  return new EventSource(url.toString(), {
    withCredentials: true,
  });
};


export const cancelStream = async (processId: string): Promise<void> => {
  await api.delete(`/ask/stream/cancel/${processId}`);
}; 