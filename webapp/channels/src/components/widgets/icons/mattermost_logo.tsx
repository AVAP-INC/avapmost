// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React from 'react';

import systemUserIcon from 'images/avapmost_system_user.png';

export default function MattermostLogo(props: React.HTMLAttributes<HTMLSpanElement>) {
    return (
        <span {...props}>
            <img
                src={systemUserIcon}
                style={{width: '100%', height: '100%', objectFit: 'cover', borderRadius: '50%'}}
                alt='Avapmost'
            />
        </span>
    );
}
