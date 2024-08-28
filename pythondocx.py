from docx import Document
from docx.enum.text import WD_ALIGN_PARAGRAPH

# document = Document("EjemploGenerales.docx")
# paragraph = document.paragraphs[0]
# print(paragraph.text)

# document = Document("EjemploGenerales.docx")
# with open('resume.xml', 'w') as f:
# 	f.write(document._element.xml)


# document = Document()
# document.add_heading('Document Title', 0)

# p = document.add_paragraph('Entre de una parte El senor ')
# p.add_run('Teodoro Alcantara Aquino ').bold = True
# p.add_run('Dominicano, mayor de edad, soltero, empleado privado, portador de la cedula de identidad y electoral No. 40221849o15, domiciliado y recidente en Santo Domingo;')
# p.add_run('quien en lo que sigue del presente contrato se denominara ')
# p.add_run('EL VENDEDOR').bold=True
# p.alignment = WD_ALIGN_PARAGRAPH.JUSTIFY
# document.save('demo.docx')



import re

def create_text_object(template):
    text_object = {
        "paragraphs": []
    }

    # Split the template into words and punctuation
    words = re.split(r"\s+", template)

    # Iterate over the words and create text nodes
    for word in words:
        match = re.match(r"{(.+):(.+)}", word)
        if match:
            # Found a curly-braced expression
            string = match.group(1)
            styles = match.group(2).split(",")
            text_object["paragraphs"].append({
                "textNode": {
                    "string": string,
                    "style": styles
                }
            })
        else:
            # Regular word or punctuation
            text_object["paragraphs"].append({
                "textNode": {
                    "string": word,
                    "style": []
                }
            })

    return text_object

# Example usage
template = "Entre de una parte {nombre:negrita,mayusculas}, de nacionalidad {nacionalidad}, mayor de edad, {estado_civil}, {ocupacion}, portador de {identificacion} No. {numero_identificacion}, domiciliado y residente en {direccion}, quien en lo que sigue del presente contrato se denominará {rol:negrita:mayusculas}, y de la otra parte: {nombre2:negrita:mayusculas}, mayor de edad, {estado_civil2}, {ocupacion}, portador de {identificaion} No. {numero_dentificacion}, domiciliado y residente en {direccion}, quien en lo que sigue del presente contrato se denominará {rol:negrita:mayuscula}, se ha convenido y pactado el siguiente"

text_object = create_text_object(template)
print(text_object)