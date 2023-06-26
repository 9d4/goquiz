const btnRefresh = document.getElementById('btn-refresh');
const theTable = document.getElementById('the-table');
const [, ...tblRows] = theTable.rows;

let pending = false;

const update = async () => {
    try {
        if (pending) return;

        pending = true;
        const res = await fetch('/adm/raw');
        const data = await res.json();

        data.students.forEach((student, i) => {
            const cellState = tblRows[i].children[2];
            const cellScore = tblRows[i].children[3];

            const elState = document.createElement('span');
            elState.classList.add('badge');
            elState.textContent = student.status;

            switch (student.status) {
                case 'Working':
                    elState.classList.add('secondary');
                    break;
                case 'Online':
                    elState.classList.add('success');
                    break;
            }

            cellState.replaceChildren(elState);
            cellScore.textContent = student.score?.value;

            const frontIndexLimit = 3;

            student.answers.forEach(answer => {
                if (answer.id == 0) {
                    return
                }

                const text = answer.correct ? '🟢' : '🔴';

                tblRows[i].children[frontIndexLimit + answer.questionId].textContent = text
            });
        });
        pending = false;
    }
    catch {
        pending = false
    }
}

btnRefresh.addEventListener('click', update);
setInterval(() => {
    update()
}, 1000);
