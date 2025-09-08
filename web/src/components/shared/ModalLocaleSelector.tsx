import { useContext, useEffect, useState } from "react";
import { LocaleContext, LocaleTypes } from "../../providers/LocaleProvider";
import React from "react";
import Button from '../shared/Button';
import { getIcon } from "./Icons";

interface Props {
  placeholder?: string;
  onClose?: () => void;

  canCloseOnBackgroundClick?: boolean;
}

export function LocalesModal(props: Props) {
  const { setLocale } = useContext(LocaleContext)
  const [isOpen, setIsOpen] = useState<boolean>(false);

  useEffect(() => {
    if (!isOpen) {
      props.onClose?.()
    }
  }, [isOpen])

  if (!isOpen) {
    return (
      <div className="cursor-default select-none flex justify-end flex-row w-min" >
        <div
          className="flex text-lg gap-2 w-min"
          onClick={() => setIsOpen((prev) => !prev)}>
          Locales
          {getIcon("Language")}
        </div>
      </div>)
  }

  return (
    <div className="cursor-default select-none flex justify-end flex-row w-min" >
      <div
        className="flex text-lg gap-2 w-min"
        onClick={() => setIsOpen((prev) => !prev)}>
        Locales
        {getIcon("Language")}
      </div>

      {isOpen && (
        <div
          className="fixed flex justify-end top-12 right-1 w-full h-full "
          onClick={() => props.canCloseOnBackgroundClick && setIsOpen(false)}
        >
          <div
            className="flex flex-col p-4 w-1/3 bg-gray-800 gap-2 h-min border-gray-500 border rounded-2xl"
            onClick={(e) => e.stopPropagation()}
          >
            {LocaleTypes.map((localeSelection) => <Button
              title={localeSelection.name}
              stretch
              onClick={() => {
                setLocale(localeSelection)
                setIsOpen(false)
              }}
            />)}
          </div>
        </div>
      )}
    </div>
  );
}

export default LocalesModal
