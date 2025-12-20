import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { userApi } from '../api';
import type { DebateSession, UserStats } from '../types';

const History: React.FC = () => {
  const [debates, setDebates] = useState<DebateSession[]>([]);
  const [stats, setStats] = useState<UserStats | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [filter, setFilter] = useState<'all' | 'wins' | 'losses' | 'draws'>('all');

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [historyData, statsData] = await Promise.all([
          userApi.getHistory(),
          userApi.getStats(),
        ]);
        setDebates(historyData || []);
        setStats(statsData);
      } catch (error) {
        console.error('Failed to fetch history:', error);
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();
  }, []);

  const getWinnerLabel = (session: DebateSession) => {
    if (session.status !== 'finished') return '⏳ 進行中';
    if (!session.winner) return '-';
    if (session.winner === 'user') return '🏆 勝利';
    if (session.winner === 'llm') return '😢 敗北';
    if (session.winner === 'draw') return '🤝 引き分け';
    if (session.winner === 'llm1') return '🤖 AI-1 勝利';
    if (session.winner === 'llm2') return '🤖 AI-2 勝利';
    return session.winner;
  };

  const filteredDebates = debates.filter((debate) => {
    if (filter === 'all') return true;
    if (filter === 'wins') return debate.winner === 'user';
    if (filter === 'losses') return debate.winner === 'llm';
    if (filter === 'draws') return debate.winner === 'draw';
    return true;
  });

  return (
    <div className="history-container">
      <header className="page-header">
        <Link to="/" className="back-link">← ホームに戻る</Link>
        <h1>📜 ディベート履歴</h1>
      </header>

      {/* 統計サマリー */}
      {stats && (
        <div className="stats-summary">
          <div className="stat-item">
            <span className="stat-value">{stats.total_debates}</span>
            <span className="stat-label">総試合</span>
          </div>
          <div className="stat-item win">
            <span className="stat-value">{stats.wins}</span>
            <span className="stat-label">勝利</span>
          </div>
          <div className="stat-item loss">
            <span className="stat-value">{stats.losses}</span>
            <span className="stat-label">敗北</span>
          </div>
          <div className="stat-item draw">
            <span className="stat-value">{stats.draws}</span>
            <span className="stat-label">引分</span>
          </div>
          <div className="stat-item">
            <span className="stat-value">{stats.win_rate.toFixed(1)}%</span>
            <span className="stat-label">勝率</span>
          </div>
        </div>
      )}

      {/* フィルター */}
      <div className="filter-buttons">
        <button
          className={`filter-btn ${filter === 'all' ? 'active' : ''}`}
          onClick={() => setFilter('all')}
        >
          すべて
        </button>
        <button
          className={`filter-btn ${filter === 'wins' ? 'active' : ''}`}
          onClick={() => setFilter('wins')}
        >
          🏆 勝利
        </button>
        <button
          className={`filter-btn ${filter === 'losses' ? 'active' : ''}`}
          onClick={() => setFilter('losses')}
        >
          😢 敗北
        </button>
        <button
          className={`filter-btn ${filter === 'draws' ? 'active' : ''}`}
          onClick={() => setFilter('draws')}
        >
          🤝 引分
        </button>
      </div>

      {/* ディベート一覧 */}
      {isLoading ? (
        <div className="loading-container">
          <div className="loading-spinner"></div>
          <p>読み込み中...</p>
        </div>
      ) : filteredDebates.length > 0 ? (
        <div className="debate-history-list">
          {filteredDebates.map((debate) => (
            <Link
              key={debate.id}
              to={`/debate/${debate.id}`}
              className="debate-history-item"
            >
              <div className="debate-main">
                <div className="debate-topic">{debate.topic}</div>
                <div className="debate-details">
                  <span className="debate-mode">
                    {debate.mode === 'user_vs_llm' ? '🤖 対AI' : '🤖vs🤖 観戦'}
                  </span>
                  {debate.mode === 'user_vs_llm' && (
                    <span className="debate-position">
                      {debate.user_position === 'pro' ? '👍 賛成側' : '👎 反対側'}
                    </span>
                  )}
                </div>
              </div>
              <div className="debate-result-section">
                <span className={`debate-result ${debate.winner}`}>
                  {getWinnerLabel(debate)}
                </span>
                <span className="debate-date">
                  {new Date(debate.created_at).toLocaleDateString('ja-JP', {
                    year: 'numeric',
                    month: 'short',
                    day: 'numeric',
                    hour: '2-digit',
                    minute: '2-digit',
                  })}
                </span>
              </div>
            </Link>
          ))}
        </div>
      ) : (
        <div className="no-data">
          <p>該当するディベートがありません</p>
          <Link to="/debate/new?mode=user_vs_llm" className="btn btn-primary">
            新しいディベートを始める
          </Link>
        </div>
      )}
    </div>
  );
};

export default History;
