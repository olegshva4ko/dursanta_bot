package tools

import "errors"

var symbols = [...]uint16{
	'`', '\'', 'ʼ',
	'а', 'А', 'б', 'Б',
	'в', 'В', 'г', 'Г',
	'ґ', 'Ґ', 'д', 'Д',
	'е', 'Е', 'є', 'Є',
	'ж', 'Ж', 'з', 'З',
	'и', 'И', 'і', 'І',
	'ї', 'Ї', 'й', 'Й',
	'к', 'К', 'л', 'Л',
	'м', 'М', 'н', 'Н',
	'о', 'О', 'п', 'П',
	'р', 'Р', 'с', 'С',
	'т', 'Т', 'у', 'У',
	'ф', 'Ф', 'х', 'Х',
	'ц', 'Ц', 'ч', 'Ч',
	'ш', 'Ш', 'щ', 'Щ',
	'ь', 'ь', 'ю', 'Ю',
	'я', 'Я',
}

//CheckName checks if message contains one of symbols
func CheckName(name []uint16) error {
M:
	for _, v := range name {
		for _, k := range symbols {
			if k == v {
				continue M
			}
		}
		return errors.New("Bad symbol")
	}
	return nil
}
