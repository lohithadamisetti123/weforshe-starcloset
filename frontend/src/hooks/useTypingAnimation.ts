import { useEffect, useState } from 'react'

interface Options {
  typingSpeed?: number
  deletingSpeed?: number
  pauseDuration?: number
}

export function useTypingAnimation(phrases: string[], options: Options = {}) {
  const { typingSpeed = 45, deletingSpeed = 25, pauseDuration = 1800 } = options
  const [phraseIndex, setPhraseIndex] = useState(0)
  const [text, setText] = useState('')
  const [isDeleting, setIsDeleting] = useState(false)

  useEffect(() => {
    const currentPhrase = phrases[phraseIndex % phrases.length]
    let timeout: ReturnType<typeof setTimeout>

    if (!isDeleting && text === currentPhrase) {
      timeout = setTimeout(() => setIsDeleting(true), pauseDuration)
    } else if (isDeleting && text === '') {
      setIsDeleting(false)
      setPhraseIndex((i) => (i + 1) % phrases.length)
    } else {
      timeout = setTimeout(
        () => {
          setText((prev) =>
            isDeleting ? currentPhrase.slice(0, prev.length - 1) : currentPhrase.slice(0, prev.length + 1)
          )
        },
        isDeleting ? deletingSpeed : typingSpeed
      )
    }

    return () => clearTimeout(timeout)
  }, [text, isDeleting, phraseIndex, phrases, typingSpeed, deletingSpeed, pauseDuration])

  return text
}
