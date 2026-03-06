import React, {useEffect, useRef, useState} from 'react';

const PLUGIN_ID = 'com.avap.avapmost-search';
const ACCENT_COLOR = '#0d5cad';

interface ReindexStatus {
    running: boolean;
    total: number;
    indexed: number;
    error: string;
}

// Props provided by Mattermost admin console custom setting machinery.
interface Props {
    id?: string;
    label?: string;
    helpText?: React.ReactNode;
    [key: string]: unknown;
}

const ReIndexButton: React.FC<Props> = () => {
    const [status, setStatus] = useState<ReindexStatus | null>(null);
    const [starting, setStarting] = useState(false);
    const [message, setMessage] = useState('');
    const pollRef = useRef<ReturnType<typeof setInterval> | null>(null);

    const stopPolling = () => {
        if (pollRef.current !== null) {
            clearInterval(pollRef.current);
            pollRef.current = null;
        }
    };

    const fetchStatus = async () => {
        try {
            const res = await fetch(`/plugins/${PLUGIN_ID}/api/v1/reindex/status`, {
                headers: {'X-Requested-With': 'XMLHttpRequest'},
                credentials: 'include',
            });
            if (res.ok) {
                const data = await res.json() as ReindexStatus;
                setStatus(data);
                if (!data.running) {
                    stopPolling();
                }
            }
        } catch {
            stopPolling();
        }
    };

    const handleStart = async () => {
        setStarting(true);
        setMessage('');
        try {
            const res = await fetch(`/plugins/${PLUGIN_ID}/api/v1/reindex/start`, {
                method: 'POST',
                headers: {'X-Requested-With': 'XMLHttpRequest'},
                credentials: 'include',
            });
            if (res.ok) {
                setMessage('再インデックスを開始しました。');
                // Start polling for status.
                stopPolling();
                pollRef.current = setInterval(() => void fetchStatus(), 2000);
                void fetchStatus();
            } else if (res.status === 409) {
                setMessage('再インデックスはすでに実行中です。');
                pollRef.current = setInterval(() => void fetchStatus(), 2000);
                void fetchStatus();
            } else if (res.status === 403) {
                setMessage('システム管理者のみ実行できます。');
            } else {
                const data = await res.json() as {error?: string};
                setMessage(`エラー: ${data.error ?? '不明なエラー'}`);
            }
        } catch {
            setMessage('通信エラーが発生しました。');
        } finally {
            setStarting(false);
        }
    };

    // Fetch initial status on mount.
    useEffect(() => {
        void fetchStatus();
        return stopPolling;
    }, []);

    const progress = status && status.total > 0
        ? Math.round((status.indexed / status.total) * 100)
        : 0;

    return (
        <div style={{padding: '8px 0'}}>
            <button
                onClick={() => void handleStart()}
                disabled={starting || (status?.running ?? false)}
                style={{
                    padding: '8px 20px',
                    borderRadius: '4px',
                    border: 'none',
                    background: (starting || status?.running) ? '#aaa' : ACCENT_COLOR,
                    color: '#fff',
                    fontSize: '14px',
                    fontWeight: 600,
                    cursor: (starting || status?.running) ? 'not-allowed' : 'pointer',
                }}
            >
                {status?.running ? '実行中...' : '全件再インデックス'}
            </button>

            {message && (
                <p style={{marginTop: '8px', fontSize: '13px', color: 'var(--center-channel-color-72, #555)'}}>
                    {message}
                </p>
            )}

            {status?.running && (
                <div style={{marginTop: '12px'}}>
                    <div style={{
                        display: 'flex', justifyContent: 'space-between',
                        fontSize: '12px', color: 'var(--center-channel-color-56, #888)',
                        marginBottom: '4px',
                    }}>
                        <span>インデックス済み: {status.indexed.toLocaleString()} / {status.total.toLocaleString()}</span>
                        <span>{progress}%</span>
                    </div>
                    <div style={{
                        width: '100%', height: '6px',
                        background: 'var(--center-channel-color-12, #e0e0e0)',
                        borderRadius: '3px', overflow: 'hidden',
                    }}>
                        <div style={{
                            height: '100%',
                            width: `${progress}%`,
                            background: ACCENT_COLOR,
                            borderRadius: '3px',
                            transition: 'width 0.3s',
                        }} />
                    </div>
                </div>
            )}

            {status != null && !status.running && status.indexed > 0 && !message && (
                <p style={{marginTop: '8px', fontSize: '13px', color: 'var(--center-channel-color-56, #888)'}}>
                    前回: {status.indexed.toLocaleString()} 件インデックス済み
                    {status.error && <span style={{color: '#e00'}}> (エラー: {status.error})</span>}
                </p>
            )}
        </div>
    );
};

export default ReIndexButton;
