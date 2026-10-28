import {InputHTMLAttributes, ReactNode} from "react";

type FieldProps = { label: string, htmlFor: string, children: ReactNode }

export function Field({label, htmlFor, children}: FieldProps) {
  return (
    <div className="flex flex-col">
      <label className="mc-label" htmlFor={htmlFor}>{label}</label>
      {children}
    </div>
  )
}

export function TextInput({className, type = "text", ...props}: InputHTMLAttributes<HTMLInputElement>) {
  return <input type={type} className={["mc-input", className].filter(Boolean).join(" ")} {...props}/>
}

type CheckboxProps = Omit<InputHTMLAttributes<HTMLInputElement>, "type"> & { label: string }

export function Checkbox({label, className, ...props}: CheckboxProps) {
  return (
    <label className={["flex cursor-pointer items-center gap-2 text-sm", className].filter(Boolean).join(" ")}>
      <span className="relative grid size-4 shrink-0">
        <input type="checkbox"
               className="peer col-start-1 row-start-1 size-4 cursor-pointer appearance-none border border-edge-control bg-inset checked:border-action-content checked:bg-action-content focus-ring"
               {...props}/>
        <svg aria-hidden="true" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="2"
             className="pointer-events-none col-start-1 row-start-1 hidden size-4 text-inset peer-checked:block">
          <path d="m4 8 3 3 5-6"/>
        </svg>
      </span>
      {label}
    </label>
  )
}
