export interface Resource {
  id: string;
  name: string;
  url: string;
  type: string;
  status: string;
  created_at: string;
  updated_at: string;
  extracted_content?: string;
  raw_content?: string;
}

export interface Notification {
  id: string;
  message: string;
  type: 'error' | 'warning' | 'info' | 'success';
  timeout?: number;
}

export interface NotificationContextType {
  notifications: Notification[];
  addNotification: (message: string, type: Notification['type'], timeout?: number) => void;
  removeNotification: (id: string) => void;
}

export interface ResourceEvent {
  id: string;
  status: string;
  error?: string;
}

export interface Message {
  role: 'user' | 'assistant';
  content: string;
  references?: Reference[];
}

export interface Reference {
  resource_id: string;
  content: string;
  score: number;
}

export interface CompleteResult {
  answer: string;
  references: Reference[];
}

export interface AskRequest {
  question: string;
}

export interface AskResponse {
  answer: string;
  references: Array<{
    resource_id: string;
    content: string;
  }>;
}

export interface SearchResult {
  answer: string;
  references: Array<{
    resource_id: string;
    content: string;
  }>;
}

export interface SaveDocumentRequest {
  name: string;
  type: string;
  content: string;
  url?: string;
} 