import './GameOverModal.css'

export default function GameOverModal({ onClose }: { onClose: () => void }) {
  return (
    <div className="game-over" role="presentation" onMouseDown={(event) => {
      if (event.target === event.currentTarget) {onClose()}
    }}>
      <section className="game-over__dialog" role="dialog" aria-modal="true" aria-labelledby="game-over-title">
        <h2 id="game-over-title">Питомец разрядился</h2>
        <p>Я полностью разрядился... 😭 Весь мой накопленный опыт стирается. Надеюсь, мы сможем еще увидеться</p>
        <button type="button" onClick={onClose} autoFocus>Понятно</button>
      </section>
    </div>
  )
}
