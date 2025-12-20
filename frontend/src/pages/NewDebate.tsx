import React, { useState } from 'react';
import { useNavigate, useSearchParams, Link } from 'react-router-dom';
import { debateApi } from '../api';
import type { DebateTopicInfo } from '../types';

const NewDebate: React.FC = () => {
  const [searchParams] = useSearchParams();
  const mode = (searchParams.get('mode') as 'user_vs_llm' | 'llm_vs_llm') || 'user_vs_llm';
  const navigate = useNavigate();

  const [topic, setTopic] = useState('');
  const [generatedTopic, setGeneratedTopic] = useState<DebateTopicInfo | null>(null);
  const [userPosition, setUserPosition] = useState<'pro' | 'con' | 'random'>('random');
  const [useRandomTopic, setUseRandomTopic] = useState(true);
  const [isLoading, setIsLoading] = useState(false);
  const [isGenerating, setIsGenerating] = useState(false);
  const [error, setError] = useState('');

  const handleGenerateTopic = async () => {
    setIsGenerating(true);
    setError('');
    try {
      const topicInfo = await debateApi.generateTopic();
      setGeneratedTopic(topicInfo);
      setTopic(topicInfo.topic);
    } catch {
      setError('テーマの生成に失敗しました');
    } finally {
      setIsGenerating(false);
    }
  };

  const handleStartDebate = async () => {
    setIsLoading(true);
    setError('');

    try {
      const response = await debateApi.createDebate({
        mode,
        topic: useRandomTopic ? '' : topic,
        user_position: mode === 'user_vs_llm' ? userPosition : undefined,
        randomize_topic: useRandomTopic,
        randomize_position: userPosition === 'random',
      });

      navigate(`/debate/${response.session.id}`, {
        state: { session: response.session, topicInfo: response.topic_info },
      });
    } catch {
      setError('ディベートの開始に失敗しました');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="new-debate-container">
      <header className="page-header">
        <Link to="/" className="back-link">← ホームに戻る</Link>
        <h1>
          {mode === 'user_vs_llm' ? '🤖 AIとディベート' : '🤖 AI同士のディベート'}
        </h1>
      </header>

      {error && <div className="error-message">{error}</div>}

      <div className="debate-setup-form">
        {/* テーマ設定 */}
        <section className="form-section">
          <h2>📋 ディベートテーマ</h2>
          
          <div className="toggle-group">
            <label className="toggle-option">
              <input
                type="radio"
                name="topicMode"
                checked={useRandomTopic}
                onChange={() => setUseRandomTopic(true)}
              />
              <span>AIにランダムで決めてもらう</span>
            </label>
            <label className="toggle-option">
              <input
                type="radio"
                name="topicMode"
                checked={!useRandomTopic}
                onChange={() => setUseRandomTopic(false)}
              />
              <span>自分で入力する</span>
            </label>
          </div>

          {!useRandomTopic && (
            <div className="topic-input-section">
              <textarea
                value={topic}
                onChange={(e) => setTopic(e.target.value)}
                placeholder="ディベートのテーマを入力してください..."
                className="topic-input"
                rows={3}
              />
              <button
                onClick={handleGenerateTopic}
                className="btn btn-secondary"
                disabled={isGenerating}
              >
                {isGenerating ? '生成中...' : '🎲 テーマを提案してもらう'}
              </button>
            </div>
          )}

          {generatedTopic && !useRandomTopic && (
            <div className="topic-preview">
              <h3>生成されたテーマ</h3>
              <p className="topic-text">{generatedTopic.topic}</p>
              <div className="positions">
                <div className="position pro">
                  <strong>賛成側:</strong> {generatedTopic.pro_position}
                </div>
                <div className="position con">
                  <strong>反対側:</strong> {generatedTopic.con_position}
                </div>
              </div>
              <p className="background">{generatedTopic.background}</p>
            </div>
          )}
        </section>

        {/* ポジション設定（ユーザー vs LLM のみ） */}
        {mode === 'user_vs_llm' && (
          <section className="form-section">
            <h2>🎭 あなたの立場</h2>
            <div className="position-options">
              <label className={`position-option ${userPosition === 'pro' ? 'selected' : ''}`}>
                <input
                  type="radio"
                  name="position"
                  value="pro"
                  checked={userPosition === 'pro'}
                  onChange={() => setUserPosition('pro')}
                />
                <span className="position-label">👍 賛成側</span>
              </label>
              <label className={`position-option ${userPosition === 'con' ? 'selected' : ''}`}>
                <input
                  type="radio"
                  name="position"
                  value="con"
                  checked={userPosition === 'con'}
                  onChange={() => setUserPosition('con')}
                />
                <span className="position-label">👎 反対側</span>
              </label>
              <label className={`position-option ${userPosition === 'random' ? 'selected' : ''}`}>
                <input
                  type="radio"
                  name="position"
                  value="random"
                  checked={userPosition === 'random'}
                  onChange={() => setUserPosition('random')}
                />
                <span className="position-label">🎲 ランダム</span>
              </label>
            </div>
          </section>
        )}

        {/* 開始ボタン */}
        <button
          onClick={handleStartDebate}
          className="btn btn-primary btn-large"
          disabled={isLoading || (!useRandomTopic && !topic.trim())}
        >
          {isLoading ? 'ディベートを準備中...' : '⚔️ ディベートを開始'}
        </button>
      </div>
    </div>
  );
};

export default NewDebate;
