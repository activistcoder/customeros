import { forwardRef } from 'react';
import { useField } from 'react-inverted-form';
import Calendar, { CalendarProps } from 'react-calendar';

type DateInputValue = null | string | number | Date;

interface DatePickerProps extends CalendarProps {
  name: string;
  formId: string;
  label?: string;

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  onBlur?: (e: any) => void;
  labelProps?: React.HTMLProps<HTMLLabelElement>;
}

export const DatePicker = forwardRef(
  (
    {
      name,
      formId,
      value,
      onBlur,
      label,
      labelProps,
      ...props
    }: DatePickerProps,
    ref,
  ) => {
    const { getInputProps } = useField(name, formId);
    const { id, onChange } = getInputProps();

    const handleDateInputChange = (data?: DateInputValue) => {
      if (!data) return onChange(null);
      const date = new Date(data);
      const normalizedDate = new Date(
        Date.UTC(date.getFullYear(), date.getMonth(), date.getDate()),
      );
      onChange(normalizedDate);
    };

    return (
      <div id={id} onBlur={onBlur}>
        <label {...labelProps}> {label} </label>
        <Calendar
          onChange={(value) => handleDateInputChange(value as DateInputValue)}
          defaultValue={value}
          ref={ref}
          {...props}
        />
      </div>
    );
  },
);