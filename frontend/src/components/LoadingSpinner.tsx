export default function LoadingSpinner({ text = 'Chargement...' }: { text?: string }) {
  return (
    <div className="loading">
      <div className="spinner" />
      <p>{text}</p>
    </div>
  )
}
