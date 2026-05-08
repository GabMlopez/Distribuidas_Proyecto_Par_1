export interface Room {
  sala_id: string;
  nombre?: string;
  tipo: 'texto' | 'multimedia';
  pin?: string;
  usuarios?: number;
}

export interface ChatMessage {
  type?: string;      
  tipo?: string;      
  texto?: string;
  nickname?: string;
  timestamp?: number;
  file_url?: string;
  file_type?: string;
}

export interface UploadTask {
  id: string;
  name: string;
  progress: number;
  done: boolean;
  error: string | null;
}
