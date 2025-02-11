// // Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// // SPDX-License-Identifier: Apache-2.0

// import { Modal } from "@cloudscape-design/components";
// import React from "react";
// import { useSelector } from 'react-redux';
// // import { useDialog } from "../../../app/hooks/useDialog";
// // import { selectVisibility } from "../../../app/reducers/dialogSlice";

// type Props = {
//     dialogKey?: string;
//     header: React.ReactNode;
//     footer?: React.ReactNode;
//     content: React.ReactNode;
//     unsavedChanges?: boolean;
//     clearInfoOnChange: () => void;
//     changeVisibility: () => void;
// };

// /**
//  *
//  * @param root0
//  * @param root0.header
//  * @param root0.footer
//  * @param root0.content
//  */
// export default function BaseDialog({ header, footer, content }: Props) {
//     const visible = useSelector(selectVisibility);

//     const { handleDismiss } = useDialog();
//     return (
//         <Modal
//             onDismiss={handleDismiss}
//             visible={visible}
//             closeAriaLabel="close"
//             header={header}
//             size="large"
//             footer={footer}>
//             {content}
//         </Modal>
//     );
// }
