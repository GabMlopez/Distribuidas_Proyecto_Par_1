export interface Room {
  sala_id: string;
  nombre?: string;
  tipo: 'texto' | 'multimedia';
  pin?: string;
  usuarios?: number;
}

export interface ChatMessage {
  tipo: 'chat' | 'join' | 'leave' | 'multimedia' | 'user_list';
  texto: string;
  nickname?: string;
  timestamp?: number;
  file_url?: string;
  users?: string[];
}

export interface UploadTask {
  id: string;
  name: string;
  progress: number;
  done: boolean;
  error: string | null;
}
