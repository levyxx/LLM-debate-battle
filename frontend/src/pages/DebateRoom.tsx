import React, { useState, useEffect, useRef } from 'react';
import { useParams, useLocation, Link, useNavigate } from 'react-router-dom';
import { debateApi } from '../api';
import type { DebateSession, DebateMessage, JudgeResult } from '../types';

const DebateRoom: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const location = useLocation();
  const navigate = useNavigate();
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const isEndingRef = useRef(false); // 審査中フラグ（ループ停止用）

  const [session, setSession] = useState<DebateSession | null>(
    location.state?.session || null
  );
  const [messages, setMessages] = useState<DebateMessage[]>([]);
  const [inputMessage, setInputMessage] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [isSending, setIsSending] = useState(false);
  const [isEnding, setIsEnding] = useState(false);
  const [judgeResult, setJudgeResult] = useState<JudgeResult | null>(null);
  const [error, setError] = useState('');
  const [isLLMDebateRunning, setIsLLMDebateRunning] = useState(false);

  // データの読み込み
  useEffect(() => {
    const fetchDebate = async () => {
      if (!id) return;
      setIsLoading(true);
      try {
        const data = await debateApi.getDebate(parseInt(id));
        setSession(data.session);
        setMessages(data.messages.filter(m => m.role !== 'system'));
        
        // 審査結果があれば取得
        const judgeMessage = data.messages.find(m => m.role === 'judge');
        if (judgeMessage) {
          try {
            setJudgeResult(JSON.parse(judgeMessage.content));
          } catch {
            // JSON解析に失敗した場合は無視
          }
        }
      } catch {
        setError('ディベートの読み込みに失敗しました');
      } finally {
        setIsLoading(false);
      }
    };

    if (!session) {
      fetchDebate();
    }
  }, [id, session]);

  // メッセージが追加されたらスクロール
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  // メッセージ送信
  const handleSendMessage = async () => {
    if (!inputMessage.trim() || !session || isSending) return;

    setIsSending(true);
    setError('');

    try {
      const response = await debateApi.sendMessage({
        session_id: session.id,
        content: inputMessage,
      });

      const newMessages: DebateMessage[] = [];
      if (response.user_message) {
        newMessages.push(response.user_message);
      }
      newMessages.push(response.llm_message);

      setMessages(prev => [...prev, ...newMessages]);
      setInputMessage('');
    } catch {
      setError('メッセージの送信に失敗しました');
    } finally {
      setIsSending(false);
    }
  };

  // LLM同士のディベートを進める
  const handleLLMDebateStep = async () => {
    if (!session || isLLMDebateRunning) return;

    setIsLLMDebateRunning(true);
    setError('');

    try {
      // 10ステップ分（5往復）を自動実行、各ステップで1つのLLMが応答
      let finished = false;
      for (let i = 0; i < 10 && !finished && !isEndingRef.current; i++) {
        const response = await debateApi.llmDebateStep(session.id);
        
        // 審査中なら即座にループを抜ける
        if (isEndingRef.current) break;

        // 応答が返ってきた瞬間に表示（1つずつ）
        if (response.llm1_message) {
          setMessages(prev => [...prev, response.llm1_message!]);
        }
        if (response.llm2_message) {
          setMessages(prev => [...prev, response.llm2_message!]);
        }

        finished = response.is_finished;

        // 次のリクエストまで少し待機
        if (!finished && !isEndingRef.current) {
          await new Promise(resolve => setTimeout(resolve, 500));
        }
      }

      // 審査中でなく、かつ終了フラグが立っている場合のみ自動審査
      if (finished && !isEndingRef.current) {
        await handleEndDebate();
      }
    } catch {
      if (!isEndingRef.current) {
        setError('ディベートの進行に失敗しました');
      }
    } finally {
      setIsLLMDebateRunning(false);
    }
  };

  // ディベート終了
  const handleEndDebate = async () => {
    if (!session || isEnding || isEndingRef.current) return;

    isEndingRef.current = true; // ループ停止フラグをセット
    setIsEnding(true);
    setError('');

    try {
      const response = await debateApi.endDebate(session.id);
      setSession(response.session);
      setJudgeResult(response.judge_result);
    } catch {
      setError('ディベートの終了に失敗しました');
    } finally {
      setIsEnding(false);
    }
  };

  // メッセージの表示スタイルを決定
  const getMessageStyle = (role: string) => {
    switch (role) {
      case 'user':
        return 'message-user';
      case 'llm':
        return 'message-llm';
      case 'llm1':
        return 'message-llm1';
      case 'llm2':
        return 'message-llm2';
      default:
        return 'message-system';
    }
  };

  // 役割のラベルを取得
  const getRoleLabel = (role: string) => {
    switch (role) {
      case 'user':
        return `あなた (${session?.user_position === 'pro' ? '賛成' : '反対'}側)`;
      case 'llm':
        return `AI (${session?.llm_position === 'pro' ? '賛成' : '反対'}側)`;
      case 'llm1':
        return 'AI-1 (賛成側)';
      case 'llm2':
        return 'AI-2 (反対側)';
      default:
        return role;
    }
  };

  if (isLoading) {
    return (
      <div className="loading-container">
        <div className="loading-spinner"></div>
        <p>読み込み中...</p>
      </div>
    );
  }

  if (!session) {
    return (
      <div className="error-container">
        <p>ディベートが見つかりません</p>
        <Link to="/" className="btn btn-primary">ホームに戻る</Link>
      </div>
    );
  }

  return (
    <div className="debate-room">
      {/* ヘッダー */}
      <header className="debate-header">
        <Link to="/" className="back-link">← ホームに戻る</Link>
        <div className="debate-info">
          <h1>{session.topic}</h1>
          <div className="debate-meta">
            <span className={`status ${session.status}`}>
              {session.status === 'ongoing' || session.status === 'active' ? '🔴 進行中' : '✅ 終了'}
            </span>
            {session.mode === 'user_vs_llm' && (
              <span className="position">
                あなた: {session.user_position === 'pro' ? '👍 賛成側' : '👎 反対側'}
              </span>
            )}
          </div>
        </div>
      </header>

      {error && <div className="error-message">{error}</div>}

      {/* メッセージエリア */}
      <div className="messages-container">
        {messages.length === 0 ? (
          <div className="empty-messages">
            <p>
              {session.mode === 'user_vs_llm'
                ? 'ディベートを開始しましょう！最初の主張を入力してください。'
                : '「ディベートを開始」ボタンをクリックしてAI同士のディベートを開始してください。'}
            </p>
          </div>
        ) : (
          messages.map((msg) => (
            <div key={msg.id} className={`message ${getMessageStyle(msg.role)}`}>
              <div className="message-header">
                <span className="message-role">{getRoleLabel(msg.role)}</span>
                <span className="message-time">
                  {new Date(msg.created_at).toLocaleTimeString('ja-JP')}
                </span>
              </div>
              <div className="message-content">{msg.content}</div>
            </div>
          ))
        )}
        <div ref={messagesEndRef} />
      </div>

      {/* 審査結果 */}
      {judgeResult && (
        <div className="judge-result">
          <h2>🏆 審査結果</h2>
          <div className="result-summary">
            <div className={`winner-badge ${judgeResult.winner}`}>
              {judgeResult.winner === 'pro' && '賛成側の勝利！'}
              {judgeResult.winner === 'con' && '反対側の勝利！'}
              {judgeResult.winner === 'draw' && '引き分け！'}
            </div>
            <div className="scores">
              <div className="score pro">
                賛成側: <strong>{judgeResult.score.pro}</strong>点
              </div>
              <div className="score con">
                反対側: <strong>{judgeResult.score.con}</strong>点
              </div>
            </div>
          </div>

          <div className="result-details">
            <h3>判定理由</h3>
            <p>{judgeResult.reasoning}</p>

            <div className="strengths-weaknesses">
              <div className="pro-feedback">
                <h4>👍 賛成側</h4>
                <div className="strengths">
                  <strong>良かった点:</strong>
                  <ul>
                    {judgeResult.pro_strengths.map((s, i) => (
                      <li key={i}>{s}</li>
                    ))}
                  </ul>
                </div>
                <div className="weaknesses">
                  <strong>改善点:</strong>
                  <ul>
                    {judgeResult.pro_weaknesses.map((w, i) => (
                      <li key={i}>{w}</li>
                    ))}
                  </ul>
                </div>
              </div>
              <div className="con-feedback">
                <h4>👎 反対側</h4>
                <div className="strengths">
                  <strong>良かった点:</strong>
                  <ul>
                    {judgeResult.con_strengths.map((s, i) => (
                      <li key={i}>{s}</li>
                    ))}
                  </ul>
                </div>
                <div className="weaknesses">
                  <strong>改善点:</strong>
                  <ul>
                    {judgeResult.con_weaknesses.map((w, i) => (
                      <li key={i}>{w}</li>
                    ))}
                  </ul>
                </div>
              </div>
            </div>

            <div className="final-comment">
              <h4>総評</h4>
              <p>{judgeResult.final_comment}</p>
            </div>
          </div>

          <button
            onClick={() => navigate('/')}
            className="btn btn-primary"
          >
            ホームに戻る
          </button>
        </div>
      )}

      {/* 入力エリア */}
      {(session.status === 'ongoing' || session.status === 'active') && !judgeResult && (
        <div className="input-area">
          {session.mode === 'user_vs_llm' ? (
            <>
              <div className="message-input-container">
                <textarea
                  value={inputMessage}
                  onChange={(e) => setInputMessage(e.target.value)}
                  placeholder="あなたの主張を入力してください..."
                  disabled={isSending}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter' && !e.shiftKey) {
                      e.preventDefault();
                      handleSendMessage();
                    }
                  }}
                />
                <button
                  onClick={handleSendMessage}
                  disabled={!inputMessage.trim() || isSending}
                  className="btn btn-primary"
                >
                  {isSending ? '送信中...' : '送信'}
                </button>
              </div>
              <div className="action-buttons">
                <button
                  onClick={handleEndDebate}
                  disabled={isEnding || messages.length < 2}
                  className="btn btn-secondary"
                >
                  {isEnding ? '審査中...' : '🏁 ディベートを終了して審査'}
                </button>
              </div>
            </>
          ) : (
            <div className="llm-debate-controls">
              {messages.length === 0 ? (
                <button
                  onClick={handleLLMDebateStep}
                  disabled={isLLMDebateRunning}
                  className="btn btn-primary btn-large"
                >
                  {isLLMDebateRunning ? 'ディベート進行中...' : '⚔️ ディベートを開始'}
                </button>
              ) : (
                <>
                  <button
                    onClick={handleLLMDebateStep}
                    disabled={isLLMDebateRunning}
                    className="btn btn-primary"
                  >
                    {isLLMDebateRunning ? '進行中...' : '▶️ 続きを見る'}
                  </button>
                  <button
                    onClick={handleEndDebate}
                    disabled={isEnding}
                    className="btn btn-secondary"
                  >
                    {isEnding ? '審査中...' : '🏁 終了して審査'}
                  </button>
                </>
              )}
            </div>
          )}
        </div>
      )}
    </div>
  );
};

export default DebateRoom;
