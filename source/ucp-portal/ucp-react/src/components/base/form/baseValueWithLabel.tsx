import { Box, BoxProps } from '@cloudscape-design/components';

interface ValueWithLabelProps {
    label: string;
    value: any;
    labelVariant?: BoxProps.Variant;
    valueVariant?: BoxProps.Variant;
}

// Context: https://cloudscape.design/components/key-value-pairs/
export default function ValueWithLabel({ label, value, labelVariant = 'h2', valueVariant = 'h3' }: ValueWithLabelProps) {
    return (
        <div>
            <Box variant={labelVariant}>{label}</Box>
            <Box variant={valueVariant} fontWeight="normal">
                {value}
            </Box>
        </div>
    );
}
