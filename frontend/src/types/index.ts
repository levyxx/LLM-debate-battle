// ユーザー
export interface User {
  id: number;
  username: string;
  created_at: string;
}

// ディベートセッション
export interface DebateSession {
  id: number;
  user_id?: number;
  topic: string;
  user_position?: string;
  llm_position?: string;
  llm1_position?: string;
  llm2_position?: string;
  mode: 'user_vs_llm' | 'llm_vs_llm';
  status: 'ongoing' | 'active' | 'finished';
  winner?: string;
  judge_comment?: string;
  created_at: string;
  finished_at?: string;
}

// ディベートメッセージ
export interface DebateMessage {
  id: number;
  session_id: number;
  role: 'user' | 'llm' | 'llm1' | 'llm2' | 'judge' | 'system';
  content: string;
  created_at: string;
}

// ディベートテーマ情報
export interface DebateTopicInfo {
  topic: string;
  pro_position: string;
  con_position: string;
  background: string;
}

// ユーザー統計
export interface UserStats {
  user_id: number;
  total_debates: number;
  wins: number;
  losses: number;
  draws: number;
  win_rate: number;
}

// 審査結果
export interface JudgeResult {
  winner: 'pro' | 'con' | 'draw';
  score: {
    pro: number;
    con: number;
  };
  reasoning: string;
  pro_strengths: string[];
  pro_weaknesses: string[];
  con_strengths: string[];
  con_weaknesses: string[];
  final_comment: string;
}

// APIリクエスト/レスポンス型
export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  token: string;
  user: User;
}

export interface CreateDebateRequest {
  mode: 'user_vs_llm' | 'llm_vs_llm';
  topic?: string;
  user_position?: 'pro' | 'con' | 'random';
  randomize_topic: boolean;
  randomize_position: boolean;
}

export interface CreateDebateResponse {
  session: DebateSession;
  topic_info?: DebateTopicInfo;
}

export interface SendMessageRequest {
  session_id: number;
  content: string;
}

export interface SendMessageResponse {
  user_message?: DebateMessage;
  llm_message: DebateMessage;
}

export interface EndDebateResponse {
  session: DebateSession;
  judge_result: JudgeResult;
}

export interface LLMDebateStepResponse {
  llm1_message?: DebateMessage;
  llm2_message?: DebateMessage;
  is_finished: boolean;
}

export interface DebateHistoryResponse {
  session: DebateSession;
  messages: DebateMessage[];
}
