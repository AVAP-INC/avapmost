import React, {useEffect, useState} from 'react';
import SearchPanel from './SearchPanel';
import ChatPanel from './ChatPanel';

const PLUGIN_ID = 'com.avap.avapmost-search';
const ACCENT_COLOR = '#0d5cad'; // Avapmost blue

type Tab = 'search' | 'chat';

interface Channel {
    id: string;
    name: string;
    display_name: string;
    type: string;
    team_id?: string;
}

interface Props {
    store?: any;
}

const RHSPanel: React.FC<Props> = ({store}) => {
    const [activeTab, setActiveTab] = useState<Tab>('search');
    const [currentChannel, setCurrentChannel] = useState<Channel | null>(null);
    const [anthropicEnabled, setAnthropicEnabled] = useState(false);

    useEffect(() => {
        fetch(`/plugins/${PLUGIN_ID}/api/v1/config`, {credentials: 'include'})
            .then((r) => r.json() as Promise<{anthropic_enabled?: boolean}>)
            .then((d) => {
                const enabled = d.anthropic_enabled ?? false;
                setAnthropicEnabled(enabled);
                if (!enabled) setActiveTab('search');
            })
            .catch(() => { /* keep false */ });
    }, []);

    useEffect(() => {
        if (!store) {
            return;
        }

        const sync = () => {
            const state = store.getState();
            const channelId = state.entities?.channels?.currentChannelId;
            const channel = channelId
                ? (state.entities?.channels?.channels?.[channelId] as Channel | undefined) ?? null
                : null;
            setCurrentChannel(channel);
        };

        sync();
        return store.subscribe(sync) as () => void;
    }, [store]);

    const channelLabel = (ch: Channel): string => {
        if (ch.display_name) return `#${ch.display_name}`;
        if (ch.type === 'D') return 'ダイレクトメッセージ';
        if (ch.type === 'G') return 'グループメッセージ';
        return `#${ch.name}`;
    };

    return (
        <div style={{display: 'flex', flexDirection: 'column', height: '100%', fontFamily: 'inherit'}}>
            {/* Header */}
            <div style={{
                padding: '10px 16px',
                borderBottom: '1px solid var(--center-channel-color-16, #e0e0e0)',
                display: 'flex',
                alignItems: 'center',
                gap: '10px',
            }}>
                <img
                    src={`/plugins/${PLUGIN_ID}/public/icon.png`}
                    alt="Search"
                    style={{height: '24px', width: '24px', objectFit: 'contain'}}
                    onError={(e) => { (e.target as HTMLImageElement).style.display = 'none'; }}
                />
                <span style={{fontWeight: 600, fontSize: '16px'}}>Avapmost Search</span>
                {currentChannel && (
                    <span style={{
                        fontSize: '12px',
                        color: 'var(--center-channel-color-56, #888)',
                        marginLeft: 'auto',
                    }}>
                        {channelLabel(currentChannel)}
                    </span>
                )}
            </div>

            {/* Tab bar */}
            <div style={{display: 'flex', borderBottom: '1px solid var(--center-channel-color-16, #e0e0e0)'}}>
                {(['search', ...(anthropicEnabled ? ['chat'] : [])] as Tab[]).map((tab) => (
                    <button
                        key={tab}
                        onClick={() => setActiveTab(tab)}
                        style={{
                            flex: 1,
                            padding: '10px',
                            border: 'none',
                            background: 'none',
                            cursor: 'pointer',
                            fontSize: '13px',
                            fontWeight: activeTab === tab ? 600 : 400,
                            borderBottom: activeTab === tab
                                ? `2px solid ${ACCENT_COLOR}`
                                : '2px solid transparent',
                            color: activeTab === tab
                                ? ACCENT_COLOR
                                : 'var(--center-channel-color, inherit)',
                        }}
                    >
                        {tab === 'search' ? '🔍 検索' : '💬 チャット'}
                    </button>
                ))}
            </div>

            {/* Panel content */}
            <div style={{flex: 1, overflow: 'hidden'}}>
                {activeTab === 'search'
                    ? <SearchPanel />
                    : <ChatPanel channel={currentChannel} />
                }
            </div>
        </div>
    );
};

export default RHSPanel;
