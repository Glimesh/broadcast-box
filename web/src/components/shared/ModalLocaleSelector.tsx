import { useContext, useEffect, useState } from "react";
import { LocaleContext, LocaleTypes } from "../../providers/LocaleProvider";
import Button from "../shared/Button";
import { getIcon } from "./Icons";

interface Props {
  placeholder?: string;
  onClose?: () => void;

  canCloseOnBackgroundClick?: boolean;
}

export function LocalesModal(props: Props) {
  const { onClose, canCloseOnBackgroundClick } = props;
  const { setLocale } = useContext(LocaleContext);
  const [isOpen, setIsOpen] = useState<boolean>(false);

  useEffect(() => {
    if (!isOpen) {
      onClose?.();
    }
  }, [isOpen, onClose]);

  return (
    <div className="relative">
      <button
        type="button"
        className="nav-link"
        onClick={() => setIsOpen((prev) => !prev)}
      >
        <span>locales</span>
        {getIcon("Language")}
      </button>

      {isOpen && (
        <>
          <div
            className="fixed inset-0 z-40"
            onClick={() => canCloseOnBackgroundClick && setIsOpen(false)}
          />
          <div className="window absolute right-0 top-[calc(100%+8px)] z-50 flex w-52 flex-col gap-1 p-2 animate-scale-in">
            {LocaleTypes.map((localeSelection) => (
              <Button
                key={localeSelection.locale}
                title={localeSelection.name}
                stretch
                classNames="justify-start"
                onClick={() => {
                  setLocale(localeSelection);
                  setIsOpen(false);
                }}
              />
            ))}
          </div>
        </>
      )}
    </div>
  );
}

export default LocalesModal;
