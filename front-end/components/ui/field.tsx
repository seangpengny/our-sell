"use client";

import { Eye, EyeOff } from "lucide-react";
import { useId, useState } from "react";
import type { InputHTMLAttributes } from "react";

type FieldProps = InputHTMLAttributes<HTMLInputElement> & {
  label: string;
  hint?: string;
  error?: string;
  trailing?: React.ReactNode;
};

export function Field({
  label,
  hint,
  error,
  trailing,
  id,
  className = "",
  "aria-describedby": describedBy,
  ...props
}: FieldProps) {
  const generatedId = useId();
  const fieldId = id ?? generatedId;
  return (
    <div className="field">
      <label className="field__label-row" htmlFor={fieldId}>
        {label}
      </label>
      <span className="field__control-wrap">
        <input
          {...props}
          id={fieldId}
          className={`field__control glass-input ${error ? "field__control--error" : ""} ${trailing ? "field__control--with-trailing" : ""} ${className}`}
          aria-invalid={Boolean(error)}
          aria-describedby={
            [
              describedBy,
              hint && `${fieldId}-hint`,
              error && `${fieldId}-error`,
            ]
              .filter(Boolean)
              .join(" ") || undefined
          }
        />
        {trailing && <span className="field__trailing">{trailing}</span>}
      </span>
      {hint && (
        <span className="field__hint" id={`${fieldId}-hint`}>
          {hint}
        </span>
      )}
      {error && (
        <span className="field__error" id={`${fieldId}-error`}>
          {error}
        </span>
      )}
    </div>
  );
}

export function PasswordField(props: Omit<FieldProps, "type" | "trailing">) {
  const [visible, setVisible] = useState(false);
  return (
    <Field
      {...props}
      type={visible ? "text" : "password"}
      trailing={
        <button
          type="button"
          className="input-icon-button"
          disabled={props.disabled}
          onClick={() => setVisible((current) => !current)}
          aria-pressed={visible}
          aria-label={visible ? "Hide password" : "Show password"}
        >
          {visible ? <EyeOff size={17} /> : <Eye size={17} />}
        </button>
      }
    />
  );
}
