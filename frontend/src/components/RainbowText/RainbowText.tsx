interface RainbowTextProps {
  text: string;
  colors?: string[];
  className?: string;
}

const RainbowText: React.FC<RainbowTextProps> = ({
  text,
  colors = ['#006398', '#00BB51', '#02D15C', '#7339C7', '#BA1A1A', '#F19B12'],
  className,
}) => {
  return (
    <span className={className} aria-label={text}>
      {text.split('').map((char, index) => (
        <span
          key={index}
          style={{ color: colors[index % colors.length] }}
        >
          {char === ' ' ? '\u00A0' : char}
        </span>
      ))}
    </span>
  );
};

export default RainbowText;