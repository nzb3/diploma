export interface Resource {
  id: string;
  name: string;
  type: string;
  status: string;
  created_at: string;
  updated_at: string;
  extracted_content?: string;
  raw_content?: string;
}

export interface ResourceEvent {
  id: string;
  status: string;
  progress: number;
  error?: string;
}

export interface SaveDocumentRequest {
  name: string;
  type: string;
  content: string;
  url?: string;
} 