/* eslint-disable react-refresh/only-export-components */
import type { PetStage } from '../../api/types'
import adultImage from '../../assets/adult.png'
import childImage from '../../assets/child.png'
import eggImage from '../../assets/egg.png'
import teenImage from '../../assets/teen.png'
import './PetVisual.css'

const stagePresentation: Record<PetStage, { label: string; image: string }> = {
  egg: { label: 'Яйцо', image: eggImage },
  child: { label: 'Малыш', image: childImage },
  teen: { label: 'Подросток', image: teenImage },
  adult: { label: 'Взрослый', image: adultImage },
}

export function stageLabel(stage: PetStage) {
  return stagePresentation[stage].label
}

export default function PetVisual({ stage }: { stage: PetStage }) {
  const presentation = stagePresentation[stage]
  return <img className="pet-visual__image" src={presentation.image} alt={`Питомец, стадия «${presentation.label}»`} />
}
