import { Calculator } from './components/Calculator/Calculator'
import styles from './App.module.css'

export default function App() {
  return (
    <div className={styles.page}>
      <header className={styles.header}>
        <h1 className={styles.title}>Calculator</h1>
        <p className={styles.subtitle}>Every result is computed by the Go service.</p>
      </header>

      <main>
        <Calculator />
      </main>

      <footer className={styles.footer}>
        <p>
          Keyboard: digits and <kbd>.</kbd>, <kbd>+</kbd> <kbd>-</kbd> <kbd>*</kbd> <kbd>/</kbd>{' '}
          <kbd>^</kbd> <kbd>%</kbd>, <kbd>r</kbd> for √, <kbd>Enter</kbd> to calculate,{' '}
          <kbd>Backspace</kbd> to delete a digit and <kbd>Esc</kbd> or <kbd>c</kbd> to clear.
        </p>
      </footer>
    </div>
  )
}
