// Компонент кнопок голосования
// TK-009/FE: Кнопки up/down для голосования, индикация текущего голоса

import { useState } from 'react';
import { voteForInitiative } from '../api';
import type { Initiative, VoteRequest } from '../types';

interface VoteButtonsProps {
  initiative: Initiative;
  onVoteChange: (updatedInitiative: Initiative) => void;
  disabled?: boolean;
}

export function VoteButtons({ initiative, onVoteChange, disabled = false }: VoteButtonsProps) {
  const [isVoting, setIsVoting] = useState(false);

  const handleVote = async (value: VoteRequest['value']) => {
    if (isVoting || disabled) return;

    // Если пользователь кликнул на уже активную кнопку, снимаем голос
    const voteValue = initiative.currentUserVote === value ? 0 : value;

    setIsVoting(true);
    try {
      const updatedInitiative = await voteForInitiative(initiative.id, { value: voteValue });
      onVoteChange(updatedInitiative);
    } catch (error) {
      console.error('Ошибка при голосовании:', error);
      // TODO: показать уведомление об ошибке
    } finally {
      setIsVoting(false);
    }
  };

  const buttonStyle = {
    background: 'transparent',
    border: '1px solid var(--color-border)',
    borderRadius: 'var(--radius)',
    padding: 'var(--space-2) var(--space-3)',
    cursor: isVoting || disabled ? 'not-allowed' : 'pointer',
    display: 'flex',
    alignItems: 'center',
    gap: 'var(--space-2)',
    fontSize: 'var(--fs-sm)',
    fontWeight: 500,
    transition: 'all 0.2s ease',
    opacity: isVoting || disabled ? 0.6 : 1,
  };

  const getUpVoteStyle = () => ({
    ...buttonStyle,
    color: initiative.currentUserVote === 1 ? 'var(--color-success)' : 'var(--color-text-muted)',
    borderColor: initiative.currentUserVote === 1 ? 'var(--color-success)' : 'var(--color-border)',
    background: initiative.currentUserVote === 1 ? 'rgba(16, 185, 129, 0.1)' : 'transparent',
  });

  const getDownVoteStyle = () => ({
    ...buttonStyle,
    color: initiative.currentUserVote === -1 ? 'var(--color-danger)' : 'var(--color-text-muted)',
    borderColor: initiative.currentUserVote === -1 ? 'var(--color-danger)' : 'var(--color-border)',
    background: initiative.currentUserVote === -1 ? 'rgba(239, 68, 68, 0.1)' : 'transparent',
  });

  return (
    <div 
      data-vote-buttons 
      style={{
        display: 'flex',
        gap: 'var(--space-2)',
        alignItems: 'center',
      }}
    >
      {/* Кнопка голоса "за" */}
      <button
        onClick={() => handleVote(1)}
        disabled={isVoting || disabled}
        style={getUpVoteStyle()}
        title={initiative.currentUserVote === 1 ? 'Снять голос' : 'Голосовать "за"'}
      >
        <span style={{ fontSize: '16px' }}>👍</span>
        <span>{initiative.upVotes}</span>
      </button>

      {/* Кнопка голоса "против" */}
      <button
        onClick={() => handleVote(-1)}
        disabled={isVoting || disabled}
        style={getDownVoteStyle()}
        title={initiative.currentUserVote === -1 ? 'Снять голос' : 'Голосовать "против"'}
      >
        <span style={{ fontSize: '16px' }}>👎</span>
        <span>{initiative.downVotes}</span>
      </button>

      {/* Счетчик общего результата */}
      <div style={{
        padding: 'var(--space-2) var(--space-3)',
        borderRadius: 'var(--radius)',
        background: 'var(--color-bg-soft)',
        color: 'var(--color-text)',
        fontSize: 'var(--fs-sm)',
        fontWeight: 600,
        minWidth: '40px',
        textAlign: 'center',
      }}>
        {initiative.voteScore > 0 && '+'}
        {initiative.voteScore}
      </div>
    </div>
  );
}
