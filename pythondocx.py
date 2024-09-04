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

sections = ["<<vendedor>>,", "<<comprador>>",
            "<<justificacion>>", "<<descripcion>>"]
data_json_simaltion = {
    "<<vendedor>>": [
        {
            "nombre": "Teodoro ",
            "apellido": "Alcantara Aquino",
            "ocupacion": "ingeniero",
            "nacionalidad": "Dominicana",
            "direccion": "C/ Antonio Maceo, edif Coronado, apto 5A, Santo Domingo, Distrino Nacional",
            "identificacion": "cedula",
            "numero_identificacion": "40221884915",
            "estado_civil": "soltero"
        },
        {
            "nombre": "Loanny",
            "apellido": "Alcantara Mojica",
            "ocupacion": "Doctora",
            "nacionalidad": "Dominicana",
            "direccion": "C/ Antonio Maceo, edif Coronado, apto 5A, Santo Domingo, Distrino Nacional",
            "identificacion": "cedula",
            "numero_identificacion": "40221884915",
            "estado_civil": "soltera"
        }
    ],
    "<<comprador>>": [
        {
            "nombre": "Anulfo",
            "apellido": "Alcantara Aquino",
            "ocupacion": "Profesor",
            "nacionalidad": "Dominicana",
            "direccion": "C/ Olmedo Paniagua, #20 Res olmedo paniagua, San Juan de la Maguana",
            "identificacion": "cedula",
            "numero_identificacion": "40251948812",
            "estado_civil": "soltero"
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
        # match2 = re.match(r"{(.+)}", word)
        if match:
            # Found a curly-braced expression
            string = match.group(1)
            styles = match.group(2).split(",")
            text_object["paragraphs"].append({
                "textNode": {
                    "string": "{"+string+'}',
                    "style": styles
                }
            })
        # elif match2:
        #     # Found a curly-braced expression
        #     string = match2.group(1)
        #     text_object["paragraphs"].append({
        #         "textNode": {
        #             "string": word,
        #             "style": []
        #         }
        #     })
        else:
            # Regular word or punctuation
            text_object["paragraphs"].append({
                "textNode": {
                    "string": word,
                    "style": []
                }
            })

    return text_object


def enhanceTemplate(main_tempate):

    template = "{nombre:negrita,mayusculas}, de nacionalidad {nacionalidad}, mayor de edad, {estado_civil}, {ocupacion}, portador de {identificacion} No. {numero_identificacion}, domiciliado y residente en {direccion}"
    templates = {
        "<<vendedor>>": template,
        "<<comprador>>": template,
        "<<descripcion>>": "Un carro",
        "<<justificacion>>": "Es mio"
    }

    for key, value in data_json_simaltion.items():
        res = ["" + templates[key]
               for r in range(len(value)) if key in templates]
        if key in templates:
            main_tempate = main_tempate.replace(key, "; ".join(res))
    return main_tempate


def buildDocument(params, template):
    # document = Document("EjemploGenerales.docx")
    # paragraph = document.paragraphs[0]
    # print(paragraph.text)

    # document = Document("EjemploGenerales.docx")
    # with open('resume.xml', 'w') as f:
    # 	f.write(document._element.xml)

    document = Document()
    document.add_heading('Document Title', 0)

    buffer = ''
    p = document.add_paragraph(buffer)
    for word in template['paragraphs']:
        w: str = ''
        w = word['textNode']['string']
        if w[0] == '{':
            p.add_run(buffer + ' ')
            replace_text_in_paragraph(word, params, p)
            buffer = ''
            # params.pop(0)
        else:
            buffer = buffer + w + ' '

    # p = document.add_paragraph('Entre de una parte El senor ')
    # p.add_run('Dominicano, mayor de edad, soltero, empleado privado, portador de la cedula de identidad y electoral No. 40221849o15, domiciliado y recidente en Santo Domingo;')
    # p.add_run('quien en lo que sigue del presente contrato se denominara ')
    # p.add_run('EL VENDEDOR').bold=True
    # p.alignment = WD_ALIGN_PARAGRAPH.JUSTIFY
    document.save('demo1.docx')


def replace_text_in_paragraph(word, params, paragraph):
    w = word['textNode']['string']
    w = w[:-1]
    w = w.replace('{', '')
    w = w.replace('}', '')
    text = ''
    if w[0] != '[':
        # TODO search the json
        text = params[0].pop(w, None) if len(params) > 0 else "Hay bobo"
        if text == None:
            params.pop(0)
            text = params[0].pop(w, None)
    else:
        text = w[1:-1]
    # fromJson = 'TEST_TEXT'
    if len(word['textNode']['style']) > 0:
        styles = word['textNode']['style']
        run = paragraph.add_run(text + ' ')
        for s in styles:
            if s == 'mayusculas':
                run.text = text.upper() + ' '
            if s == 'negrita':
                run.bold = True
    else:
        paragraph.add_run(text + '')


# Example usage
template2 = "Entre de una parte <<vendedor>>, quien en lo que sigue del presente contrato se denominará {[vendedor]:negrita,mayusculas}, y de la otra parte: <<comprador>>, quien en lo que sigue del presente contrato se denominará {[comprador]:negrita,mayusculas}, se ha convenido y pactado el siguiente"

template2 = enhanceTemplate(template2)

text_object = create_text_object(template2)
print(text_object)

params = []
for key, value in data_json_simaltion.items():
    params.append(value)

params = [item for values in params for item in values]
buildDocument(params, text_object)
