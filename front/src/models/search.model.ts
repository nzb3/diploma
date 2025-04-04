export interface AskRequest {
  question: string;
}

export interface Reference {
  resource_id: string;
  content: string;
  score: number;
}

export interface AskResponse {
  answer: string;
  references: Reference[];
}

export interface SearchResult {
  answer: string;
  references: Reference[];
}

export interface Message {
  role: 'user' | 'assistant';
  content: string;
}

export interface CompleteResult {
  answer: string;
  references: Reference[];
} 