interface notificationProps {
  message: string
}

export function Notification({message}: notificationProps) {

  return (
    <div className="mc-notification">
      <span>{message}</span>
    </div>)
}
