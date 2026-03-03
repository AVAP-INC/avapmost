// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React from 'react';

import logoImage from 'images/logo.svg';

type Props = {
    width?: number;
    height?: number;
    className?: string;
}

export default (props: Props) => (
    <img
        className={props.className}
        src={logoImage}
        height={props.height || 30}
        style={{width: 'auto'}}
        alt='Avapmost'
    />
);
