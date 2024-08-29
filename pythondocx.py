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

sections = ["<<vendedor>>,", "<<comprador>>", "<<justificacion>>", "<<descripcion>>"]

data_json_simaltion = {
    "<<vendedor>>": [
        {
            "nombre": "Teodoro ",
            "apellido": "Alcantara Aquino",
            "ocupacion": "ingeniero",
            "nacionalidad": "Dominicana",
            "direccion": "C/ Antonio Maceo, edif Coronado, apto 5A, Santo Domingo, Distrino Nacional",
            "identificacion": "cedula",
            "no_identificacion": "40221884915"
        },
        {
            "nombre": "Loanny",
            "apellido": "Alcantara Mojica",
            "ocupacion": "Doctora",
            "nacionalidad": "Dominicana",
            "direccion": "C/ Antonio Maceo, edif Coronado, apto 5A, Santo Domingo, Distrino Nacional",
            "identificacion": "cedula",
            "no_identificacion": "40221884915"
        }
    ],
    "comprador": [
        {
            "nombre": "Anulfo",
            "apellido": "Alcantara Aquino",
            "ocupacion": "Profesor",
            "nacionalidad": "Dominicana",
            "direccion": "C/ Olmedo Paniagua, #20 Res olmedo paniagua, San Juan de la Maguana",
            "identificacion": "cedula",
            "no_identificacion": "40251948812"
        }
    ],
    "<<descripcion>>": ["Un carro"],
    "<<justificacion>>": ["Es mio"]
}


def create_text_object(template):
    template2 = "{nombre:negrita,mayusculas}, de nacionalidad {nacionalidad}, mayor de edad, {estado_civil}, {ocupacion}, portador de {identificacion} No. {numero_identificacion}, domiciliado y residente en {direccion}"
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
template = "{nombre:negrita,mayusculas}, de nacionalidad {nacionalidad}, mayor de edad, {estado_civil}, {ocupacion}, portador de {identificacion} No. {numero_identificacion}, domiciliado y residente en {direccion}"
template2 = "Entre de una parte <<vendedor>>, quien en lo que sigue del presente contrato se denominará {vendedor:negrita:mayusculas}, y de la otra parte: <<comprador>>, quien en lo que sigue del presente contrato se denominará {comprador:negrita:mayusculas}, se ha convenido y pactado el siguiente"

templates = {
    "<<vendedor>>": template,
    "<<comprador>>": template,
    "<<descripcion>>": "Un carro",
    "<<justificacion>>": "Es mio"
}

for key, value in data_json_simaltion.items():
    res = ["" + templates[key]  for r in range(len(value)) if key in templates]
    if key in templates:
        template2 = template2.replace(key, "; ".join(res))
text_object = create_text_object(template2)
print(text_object)
