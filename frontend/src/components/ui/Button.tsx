import { motion } from "motion/react";
import type React from "react";

type ButtonVariant = "primary" | "accent" | "secondary" | "outline" | "ghost";

interface ButtonProps {
  variant?: ButtonVariant;
  children: React.ReactNode;
  className?: string;
  disabled?: boolean;
  onClick?: (e: React.MouseEvent<HTMLButtonElement>) => void;
  type?: "button" | "submit" | "reset";
  "aria-label"?: string;
}

const variantClasses: Record<ButtonVariant, string> = {
  primary: "bg-primary text-primary-content hover:brightness-110",
  accent: "bg-accent text-accent-content hover:brightness-110",
  secondary: "bg-soft-brown text-text-brown hover:brightness-95",
  outline: "border-2 border-primary text-primary bg-transparent hover:bg-primary/10",
  ghost: "bg-cream text-text-brown border border-soft-brown/50 hover:bg-soft-brown/30",
};

export function Button({
  variant = "primary",
  className = "",
  children,
  ...props
}: ButtonProps): React.JSX.Element {
  return (
    <motion.button
      className={`rounded-[8px] px-5 py-2 text-sm font-medium cursor-pointer transition-colors duration-150 ${variantClasses[variant]} ${className}`}
      whileHover={{ scale: 1.04 }}
      whileTap={{ scale: 0.97 }}
      {...props}
    >
      {children}
    </motion.button>
  );
}
