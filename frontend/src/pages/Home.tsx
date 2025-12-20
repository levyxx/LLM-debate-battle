import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { userApi } from '../api';
import type { UserStats, DebateSession } from '../types';

const Home: React.FC = () => {
  const { user, logout } = useAuth();
  const [stats, setStats] = useState<UserStats | null>(null);
  const [recentDebates, setRecentDebates] = useState<DebateSession[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [statsData, historyData] = await Promise.all([
          userApi.getStats(),
          userApi.getHistory(),
        ]);
        setStats(statsData);
        setRecentDebates(historyData.slice(0, 5));
      } catch (error) {
        console.error('Failed to fetch data:', error);
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();
  }, []);

  const getWinnerLabel = (session: DebateSession) => {
    if (!session.winner) return '-';
    if (session.winner === 'user') return '🏆 勝利';
    if (session.winner === 'llm') return '😢 敗北';
    if (session.winner === 'draw') return '🤝 引き分け';
    return session.winner;
  };

  return (
    <div className="home-container">
      <header className="home-header">
        <h1>🎯 LLMディベートバトル</h1>
        <div className="user-info">
          <span>ようこそ、{user?.username}さん</span>
          <button onClick={logout} className="btn btn-secondary btn-small">
            ログアウト
          </button>
        </div>
      </header>

      <main className="home-main">
        {/* 統計カード */}
        <section className="stats-section">
          <h2>📊 あなたの戦績</h2>
          {isLoading ? (
            <p>読み込み中...</p>
          ) : stats ? (
            <div className="stats-grid">
              <div className="stat-card">
                <div className="stat-value">{stats.total_debates}</div>
                <div className="stat-label">総ディベート数</div>
              </div>
              <div className="stat-card win">
                <div className="stat-value">{stats.wins}</div>
                <div className="stat-label">勝利</div>
              </div>
              <div className="stat-card loss">
                <div className="stat-value">{stats.losses}</div>
                <div className="stat-label">敗北</div>
              </div>
              <div className="stat-card draw">
                <div className="stat-value">{stats.draws}</div>
                <div className="stat-label">引き分け</div>
              </div>
              <div className="stat-card">
                <div className="stat-value">{stats.win_rate.toFixed(1)}%</div>
                <div className="stat-label">勝率</div>
              </div>
            </div>
          ) : (
            <p>統計データがありません</p>
          )}
        </section>

        {/* アクションボタン */}
        <section className="action-section">
          <h2>🎮 新しいディベートを始める</h2>
          <div className="action-buttons">
            <Link to="/debate/new?mode=user_vs_llm" className="action-card">
              <div className="action-icon">🤖</div>
              <h3>AIとディベート</h3>
              <p>AIを相手にディベートで勝負！</p>
            </Link>
            <Link to="/debate/new?mode=llm_vs_llm" className="action-card">
              <div className="action-icon">🤖 vs 🤖</div>
              <h3>AI同士のディベート</h3>
              <p>AIが議論する様子を観戦！</p>
            </Link>
          </div>
        </section>

        {/* 最近のディベート */}
        <section className="history-section">
          <div className="section-header">
            <h2>📜 最近のディベート</h2>
            <Link to="/history" className="btn btn-link">
              すべて見る →
            </Link>
          </div>
          {isLoading ? (
            <p>読み込み中...</p>
          ) : recentDebates.length > 0 ? (
            <div className="debate-list">
              {recentDebates.map((debate) => (
                <Link
                  key={debate.id}
                  to={`/debate/${debate.id}`}
                  className="debate-list-item"
                >
                  <div className="debate-topic">{debate.topic}</div>
                  <div className="debate-meta">
                    <span className="debate-mode">
                      {debate.mode === 'user_vs_llm' ? '対AI' : 'AI観戦'}
                    </span>
                    <span className="debate-result">{getWinnerLabel(debate)}</span>
                    <span className="debate-date">
                      {new Date(debate.created_at).toLocaleDateString('ja-JP')}
                    </span>
                  </div>
                </Link>
              ))}
            </div>
          ) : (
            <p className="no-data">まだディベート履歴がありません</p>
          )}
        </section>
      </main>
    </div>
  );
};

export default Home;
