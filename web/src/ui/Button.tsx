import {ButtonHTMLAttributes} from "react";
import {Link, LinkProps} from "react-router-dom";

export type ButtonVariant = "primary" | "secondary" | "ghost"
export type ButtonSize = "md" | "sm" | "icon"

const variants: Record<ButtonVariant, string> = {
  primary: "mc-button",
  secondary: "mc-button-secondary",
  ghost: "mc-button-ghost",
}

const sizes: Record<ButtonVariant, Record<ButtonSize, string>> = {
  primary: {md: "", sm: "px-3 py-1 text-sm", icon: "px-3"},
  secondary: {md: "", sm: "px-3 py-1 text-sm", icon: "px-3"},
  ghost: {md: "", sm: "px-2 py-1", icon: "px-2"},
}

function buttonClassName(variant: ButtonVariant = "primary", size: ButtonSize = "md", className = "") {
  return [variants[variant], sizes[variant][size], className].filter(Boolean).join(" ")
}

type Props = ButtonHTMLAttributes<HTMLButtonElement> & { variant?: ButtonVariant, size?: ButtonSize }

export function Button({variant = "primary", size = "md", className, type = "button", ...props}: Props) {
  return <button type={type} className={buttonClassName(variant, size, className)} {...props}/>
}

type LinkButtonProps = LinkProps & { variant?: ButtonVariant, size?: ButtonSize }

export function LinkButton({variant = "primary", size = "md", className, ...props}: LinkButtonProps) {
  return <Link className={buttonClassName(variant, size, className)} {...props}/>
}
