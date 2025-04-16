import axios from 'axios';
import {AskRequest, AskResponse, Resource, SaveDocumentRequest} from '../types/api';

const API_BASE_URL = '/api/v1';

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
  
  const eventSourceUrl = `${window.location.origin}${API_BASE_URL}/resources`;

  return new EventSource(eventSourceUrl, {
    withCredentials: true,
  });
};

export const askQuestion = async (data: AskRequest): Promise<AskResponse> => {
  const response = await api.post('/ask', data);
  return response.data;
};

export const streamAnswer = (question: string): EventSource => {
  const url = new URL(`${window.location.origin}${API_BASE_URL}/ask/stream`);

  url.searchParams.append('question', question);

  return new EventSource(url.toString(), {
    withCredentials: true,
  });
};


export const cancelStream = async (processId: string): Promise<void> => {
  await api.delete(`/ask/stream/cancel/${processId}`);
}; 