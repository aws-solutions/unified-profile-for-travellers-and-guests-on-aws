// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import { Button } from '@cloudscape-design/components';
import { DeleteErrorRequest } from '../../models/error';
import { useDeleteErrorMutation } from '../../store/errorApiSlice';
import { IconName } from '../../models/iconName';

interface DeleteErrorButtonProps {
    errorId: string;
}

export const DeleteErrorButton = (props: DeleteErrorButtonProps) => {
    const deleteErrorRequest: DeleteErrorRequest = {
        errorId: props.errorId,
    };

    const [deleteErrorTrigger] = useDeleteErrorMutation();
    const onDeleteErrorButtonClick = async (): Promise<void> => {
        deleteErrorTrigger({ deleteErrorRequest });
    };

    return <Button variant="icon" onClick={onDeleteErrorButtonClick} iconName={IconName.REMOVE} />;
};
