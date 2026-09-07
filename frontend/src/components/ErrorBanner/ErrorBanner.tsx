import styles from './ErrorBanner.module.css'

interface ErrorBannerProps {
  message: string | null
}

export function ErrorBanner({ message }: ErrorBannerProps) {
  return (
    <p className={styles.banner} role="alert" hidden={message === null}>
      {message}
    </p>
  )
}
