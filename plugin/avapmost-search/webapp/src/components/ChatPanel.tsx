import React, {useEffect, useRef, useState} from 'react';

const PLUGIN_ID = 'com.avap.avapmost-search';
const ACCENT_COLOR = '#0d5cad'; // Avapmost blue

type Target = 'channel' | 'team' | 'all';

const TARGET_OPTIONS: {id: Target; label: string; title: string}[] = [
    {id: 'channel', label: 'チャンネル', title: '現在のチャンネルのみ'},
    {id: 'team',    label: 'チーム',     title: 'チーム内の全チャンネル'},
    {id: 'all',     label: '全体',       title: '参加している全チャンネル'},
];

interface Message {
    role: 'user' | 'bot';
    text: string;
}

interface Channel {
    id: string;
    name: string;
    display_name: string;
    type: string;
    team_id?: string;
}

interface Props {
    channel?: Channel | null;
}

const channelLabel = (ch: Channel): string => {
    if (ch.display_name) return `#${ch.display_name}`;
    if (ch.type === 'D') return 'ダイレクトメッセージ';
    if (ch.type === 'G') return 'グループメッセージ';
    return `#${ch.name}`;
};

const TargetSelector: React.FC<{target: Target; onChange(t: Target): void}> = ({target, onChange}) => (
    <div style={{display: 'flex', gap: '4px', alignItems: 'center'}}>
        {TARGET_OPTIONS.map(({id, label, title}) => (
            <button
                key={id}
                title={title}
                onClick={() => onChange(id)}
                style={{
                    padding: '3px 8px',
                    border: `1px solid ${target === id ? 'var(--button-bg, #166de0)' : 'var(--center-channel-color-24, #ccc)'}`,
                    borderRadius: '3px',
                    background: target === id ? 'var(--button-bg, #166de0)' : 'transparent',
                    color: target === id ? '#fff' : 'var(--center-channel-color-56, #888)',
                    fontSize: '11px',
                    fontWeight: target === id ? 600 : 400,
                    cursor: 'pointer',
                    lineHeight: 1.4,
                    transition: 'all 0.1s',
                }}
            >
                {label}
            </button>
        ))}
    </div>
);

const greeting = (target: Target, channel: Channel | null | undefined): string => {
    if (target === 'channel' && channel) {
        return `こんにちは！Avapmost Search です。\n${channelLabel(channel)} について何でも聞いてください 🤖`;
    }
    if (target === 'team') {
        return 'こんにちは！Avapmost Search です。\nチーム全体のメッセージを参照して回答します 🤖';
    }
    return 'こんにちは！Avapmost Search です。\n参加している全チャンネルを参照して回答します 🤖';
};

const ChatPanel: React.FC<Props> = ({channel}) => {
    const [target, setTarget] = useState<Target>('channel');
    const [messages, setMessages] = useState<Message[]>([
        {role: 'bot', text: greeting('channel', channel)},
    ]);
    const [input, setInput] = useState('');
    const [loading, setLoading] = useState(false);
    const bottomRef = useRef<HTMLDivElement>(null);

    // チャンネルまたはターゲットが切り替わったら挨拶を更新
    useEffect(() => {
        setMessages([{role: 'bot', text: greeting(target, channel)}]);
    }, [channel?.id, target]);

    useEffect(() => {
        bottomRef.current?.scrollIntoView({behavior: 'smooth'});
    }, [messages]);

    const handleSend = async () => {
        const question = input.trim();
        if (!question || loading) return;
        setInput('');
        setMessages((prev) => [...prev, {role: 'user', text: question}]);
        setLoading(true);
        try {
            const res = await fetch(`/plugins/${PLUGIN_ID}/api/v1/ask`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'X-Requested-With': 'XMLHttpRequest',
                },
                credentials: 'include',
                body: JSON.stringify({
                    question,
                    target,
                    context: channel ? {
                        channel_id: channel.id,
                        channel_name: channel.display_name || channel.name,
                        channel_type: channel.type,
                        team_id: channel.team_id ?? '',
                    } : undefined,
                }),
            });
            if (res.ok) {
                const data = await res.json() as {answer?: string; error?: string};
                if (data.error) {
                    setMessages((prev) => [...prev, {role: 'bot', text: `エラー: ${data.error}`}]);
                } else {
                    setMessages((prev) => [...prev, {role: 'bot', text: data.answer ?? '…'}]);
                }
            } else {
                setMessages((prev) => [...prev, {
                    role: 'bot',
                    text: '申し訳ありません。応答を取得できませんでした。',
                }]);
            }
        } catch {
            setMessages((prev) => [...prev, {role: 'bot', text: '通信エラーが発生しました。'}]);
        } finally {
            setLoading(false);
        }
    };

    return (
        <div style={{
            display: 'flex',
            flexDirection: 'column',
            height: '100%',
            padding: '8px',
            boxSizing: 'border-box',
        }}>
            {/* Target selector */}
            <div style={{
                display: 'flex',
                alignItems: 'center',
                gap: '8px',
                paddingBottom: '8px',
                borderBottom: '1px solid var(--center-channel-color-12, rgba(0,0,0,0.08))',
                marginBottom: '4px',
            }}>
                <span style={{fontSize: '11px', color: 'var(--center-channel-color-56, #888)'}}>スコープ:</span>
                <TargetSelector target={target} onChange={setTarget} />
            </div>

            <div style={{flex: 1, overflowY: 'auto', padding: '4px'}}>
                {messages.map((msg, i) => (
                    <div
                        key={i}
                        style={{
                            display: 'flex',
                            justifyContent: msg.role === 'user' ? 'flex-end' : 'flex-start',
                            alignItems: 'flex-end',
                            gap: '6px',
                            marginBottom: '8px',
                        }}
                    >
                        {msg.role === 'bot' && (
                            <img
                                src={`/plugins/${PLUGIN_ID}/public/avapmost_system_user.png`}
                                alt="Avapmost Search"
                                style={{width: '28px', height: '28px', borderRadius: '50%', flexShrink: 0, objectFit: 'cover'}}
                            />
                        )}
                        <div style={{
                            maxWidth: '75%',
                            padding: '8px 12px',
                            borderRadius: msg.role === 'user'
                                ? '12px 12px 4px 12px'
                                : '12px 12px 12px 4px',
                            background: msg.role === 'user'
                                ? ACCENT_COLOR
                                : 'var(--center-channel-color-08, #f0f0f0)',
                            color: msg.role === 'user'
                                ? '#fff'
                                : 'var(--center-channel-color, #000)',
                            fontSize: '13px',
                            lineHeight: '1.6',
                            whiteSpace: 'pre-wrap',
                        }}>
                            {msg.text}
                        </div>
                    </div>
                ))}

                {loading && (
                    <div style={{display: 'flex', justifyContent: 'flex-start', alignItems: 'flex-end', gap: '6px', marginBottom: '8px'}}>
                        <img
                            src={`/plugins/${PLUGIN_ID}/public/avapmost_system_user.png`}
                            alt="Avapmost Search"
                            style={{width: '28px', height: '28px', borderRadius: '50%', flexShrink: 0, objectFit: 'cover'}}
                        />
                        <div style={{
                            padding: '8px 12px',
                            borderRadius: '12px 12px 12px 4px',
                            background: 'var(--center-channel-color-08, #f0f0f0)',
                            fontSize: '13px',
                            color: 'var(--center-channel-color-56, #888)',
                        }}>
                            考え中…
                        </div>
                    </div>
                )}

                <div ref={bottomRef} />
            </div>

            <div style={{
                display: 'flex',
                gap: '8px',
                paddingTop: '8px',
                borderTop: '1px solid var(--center-channel-color-16, #e0e0e0)',
            }}>
                <input
                    type="text"
                    value={input}
                    onChange={(e) => setInput(e.target.value)}
                    onKeyDown={(e) => {
                        if (e.key === 'Enter' && !e.shiftKey) {
                            void handleSend();
                        }
                    }}
                    placeholder="メッセージを入力… (Enter で送信)"
                    disabled={loading}
                    style={{
                        flex: 1,
                        padding: '8px 12px',
                        borderRadius: '4px',
                        border: '1px solid var(--center-channel-color-24, #ccc)',
                        background: 'var(--center-channel-bg, #fff)',
                        color: 'var(--center-channel-color, #000)',
                        fontSize: '14px',
                        outline: 'none',
                    }}
                />
                <button
                    onClick={() => void handleSend()}
                    disabled={loading || !input.trim()}
                    style={{
                        padding: '8px 16px',
                        borderRadius: '4px',
                        border: 'none',
                        background: ACCENT_COLOR,
                        color: '#fff',
                        cursor: loading || !input.trim() ? 'not-allowed' : 'pointer',
                        fontSize: '14px',
                        opacity: !input.trim() ? 0.6 : 1,
                    }}
                >
                    送信
                </button>
            </div>
        </div>
    );
};

export default ChatPanel;
