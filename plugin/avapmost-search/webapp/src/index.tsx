import React from 'react';
import RHSPanel from './components/RHSPanel';
import ReIndexButton from './components/ReIndexButton';

const PLUGIN_ID = 'com.avap.avapmost-search';

class AvapmostSearchPlugin {
    initialize(registry: any, store: any): void {
        const {toggleRHSPlugin} = registry.registerRightHandSidebarComponent(
            // eslint-disable-next-line react/display-name
            () => React.createElement(RHSPanel, {store} as React.ComponentProps<typeof RHSPanel>),
            'Avapmost Search',
        );

        registry.registerAppBarComponent(
            `/plugins/${PLUGIN_ID}/public/icon.png`,
            () => store.dispatch(toggleRHSPlugin),
            'Avapmost Search',
        );

        // Register the reindex button in the admin console custom setting.
        registry.registerAdminConsoleCustomSetting('ReIndexSection', ReIndexButton, {showTitle: false});
    }

    uninitialize(): void {
        // nothing to clean up
    }
}

declare global {
    interface Window {
        registerPlugin(id: string, plugin: AvapmostSearchPlugin): void;
    }
}

window.registerPlugin(PLUGIN_ID, new AvapmostSearchPlugin());
