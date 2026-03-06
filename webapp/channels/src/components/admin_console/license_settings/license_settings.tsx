// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React from 'react';
import {FormattedMessage, defineMessages} from 'react-intl';

import type {ClientLicense, EnvironmentConfig} from '@mattermost/types/config';
import type {ServerError} from '@mattermost/types/errors';
import type {ServerLimits} from '@mattermost/types/limits';
import type {GetFilteredUsersStatsOpts, UsersStats} from '@mattermost/types/users';
import type {StatusOK} from '@mattermost/types/client4';

import type {ActionResult} from 'mattermost-redux/types/actions';

import AdminHeader from 'components/widgets/admin_console/admin_header';

import type {ModalData} from 'types/actions';

import './license_settings.scss';

type Props = {
    license: ClientLicense;
    enterpriseReady: boolean;
    upgradedFromTE: boolean;
    totalUsers: number;
    isDisabled: boolean;
    prevTrialLicense: ClientLicense;
    environmentConfig: Partial<EnvironmentConfig>;
    actions: {
        getLicenseConfig: () => void;
        uploadLicense: (file: File) => Promise<ActionResult>;
        removeLicense: () => Promise<ActionResult<boolean, ServerError>>;
        getPrevTrialLicense: () => void;
        upgradeToE0: () => Promise<StatusOK>;
        upgradeToE0Status: () => Promise<{percentage: number; error: string | JSX.Element | null}>;
        isAllowedToUpgradeToEnterprise: () => Promise<ActionResult>;
        restartServer: () => Promise<StatusOK>;
        ping: () => Promise<{status: string}>;
        requestTrialLicense: (users: number, termsAccepted: boolean, receiveEmailsAccepted: boolean, featureName: string) => Promise<ActionResult>;
        openModal: <P>(modalData: ModalData<P>) => void;
        getServerLimits: () => Promise<ActionResult<ServerLimits, ServerError>>;
        getFilteredUsersStats: (filters: GetFilteredUsersStatsOpts) => Promise<{
            data?: UsersStats;
            error?: ServerError;
        }>;
    };
}

const messages = defineMessages({
    title: {id: 'admin.license.title', defaultMessage: 'Edition and License'},
});

export const searchableStrings = [
    messages.title,
];

export default class LicenseSettings extends React.PureComponent<Props> {
    componentDidMount() {
        this.props.actions.getLicenseConfig();
    }

    render() {
        const {license} = this.props;
        const currentYear = new Date().getFullYear();

        return (
            <div className='wrapper--fixed'>
                <AdminHeader>
                    <FormattedMessage {...messages.title}/>
                </AdminHeader>
                <div className='admin-console__wrapper'>
                    <div className='admin-console__content'>
                        <div className='top-wrapper'>
                            <div className='left-panel'>
                                <div className='panel-card'>
                                    {/* NOTE: Do not localise these strings. Legally we cannot since the license is in English. */}
                                    <h3 style={{marginTop: 0}}>{'Avapmost Edition'}</h3>
                                    <p>
                                        {'Avapmost is a fork of '}
                                        <strong>{'Mattermost®'}</strong>
                                        {', an open source messaging platform developed by Mattermost, Inc. This fork is developed and maintained by AVAP-INC.'}
                                    </p>
                                    <p>
                                        {'© '}
                                        {currentYear}
                                        {' AVAP Co.,Ltd | Based on Mattermost © Mattermost, Inc.'}
                                    </p>
                                    <hr style={{borderColor: 'rgba(var(--center-channel-color-rgb), 0.12)', margin: '16px 0'}}/>
                                    <h4 style={{marginTop: 0}}>{'Mattermost License'}</h4>
                                    <p>{'When using Mattermost Team Edition, the software is offered under a Mattermost MIT Compiled License. See MIT-COMPILED-LICENSE.md in your root install directory for details.'}</p>
                                    <p>{'When using Mattermost Enterprise Edition, the software is offered under a commercial license. See the Enterprise Edition License for details.'}</p>
                                    <p>{'See NOTICE.txt for information about open source software used in the system.'}</p>
                                </div>
                            </div>
                            <div className='right-panel'>
                                <div className='panel-card'>
                                    {/* NOTE: Do not localise these strings. Legally we cannot since the license is in English. */}
                                    <h4 style={{marginTop: 0}}>
                                        {license.IsLicensed === 'true' ? 'Licensed' : 'Unlicensed (Team Edition)'}
                                    </h4>
                                    {license.IsLicensed === 'true' ? (
                                        <div>
                                            {license.SkuName && <p><strong>{'SKU: '}</strong>{license.SkuName}</p>}
                                            {license.Company && <p><strong>{'Company: '}</strong>{license.Company}</p>}
                                            {license.Users && <p><strong>{'Licensed users: '}</strong>{license.Users}</p>}
                                            {license.IssuedAt && (
                                                <p>
                                                    <strong>{'Issued: '}</strong>
                                                    {new Date(parseInt(license.IssuedAt, 10)).toLocaleDateString()}
                                                </p>
                                            )}
                                            {license.ExpiresAt && (
                                                <p>
                                                    <strong>{'Expires: '}</strong>
                                                    {new Date(parseInt(license.ExpiresAt, 10)).toLocaleDateString()}
                                                </p>
                                            )}
                                        </div>
                                    ) : (
                                        <p>{'No license key loaded. Running as Team Edition.'}</p>
                                    )}
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        );
    }
}
