import { Eye, EyeOff } from "lucide-react";
import { useId, useState } from "react";
import type { InputHTMLAttributes } from "react";

type FieldProps = InputHTMLAttributes<HTMLInputElement> & {
  label: string;
  hint?: string;
  error?: string;
  trailing?: React.ReactNode;
};

export function Field({ label, hint, error, trailing, id, className = "", ...props }: FieldProps) {
  const generatedId = useId();
  const fieldId = id ?? generatedId;
  return (
    <label className="field" htmlFor={fieldId}>
      <span className="field__label-row"><span>{label}</span>{hint && <span className="field__hint">{hint}</span>}</span>
      <span className="field__control-wrap">
        <input id={fieldId} className={`field__control ${error ? "field__control--error" : ""} ${className}`} aria-invalid={Boolean(error)} {...props} />
        {trailing && <span className="field__trailing">{trailing}</span>}
      </span>
      {error && <span className="field__error">{error}</span>}
    </label>
  );
}

export function PasswordField(props: Omit<FieldProps, "type" | "trailing">) {
  const [visible, setVisible] = useState(false);
  return <Field {...props} type={visible ? "text" : "password"} trailing={<button type="button" className="input-icon-button" onClick={() => setVisible((current) => !current)} aria-label={visible ? "Hide password" : "Show password"}>{visible ? <EyeOff size={17} /> : <Eye size={17} />}</button>} />;
}
