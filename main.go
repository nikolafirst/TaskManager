package main

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Task struct {
	Id          uint
	Title       string
	Description string
}

func main() {
	a := app.New()
	a.Settings().SetTheme(theme.LightTheme())

	w := a.NewWindow("Менеджер Задач")
	w.Resize(fyne.NewSize(500, 400))
	w.CenterOnScreen()

	var tasks []Task
	var createContent *fyne.Container
	var tasksContent *fyne.Container
	var editContent *fyne.Container
	var tasksList *widget.List

	db, _ := gorm.Open(sqlite.Open("tasks.db"), &gorm.Config{})
	db.AutoMigrate(&Task{})
	db.Find(&tasks)

	noTasksLabel := canvas.NewText("Нет задач ...", color.Black)
	if len(tasks) != 0 {
		noTasksLabel.Hide()
	}

	newTaskIcon, _ := fyne.LoadResourceFromPath("./icons/new.png")
	back, _ := fyne.LoadResourceFromPath("./icons/back.png")
	save, _ := fyne.LoadResourceFromPath("./icons/save.png")
	delete, _ := fyne.LoadResourceFromPath("./icons/delete.png")
	edit, _ := fyne.LoadResourceFromPath("./icons/edit.png")

	taskBar := container.NewHBox(
		canvas.NewText("Ваши задачи", color.Black),
		layout.NewSpacer(),
		widget.NewButtonWithIcon(
			"",
			newTaskIcon,
			func() {
				w.SetContent(createContent)
			}),
	)

	tasksList = widget.NewList(
		func() int {
			return len(tasks)
		},

		func() fyne.CanvasObject {
			return widget.NewLabel("Default text")
		},

		func(lii widget.ListItemID, co fyne.CanvasObject) {
			co.(*widget.Label).SetText(tasks[lii].Title)
		},
	)

	// Выбор таска из списка

	tasksList.OnSelected = func(id widget.ListItemID) {
		detailsBar := container.NewHBox(
			canvas.NewText(
				fmt.Sprintf(
					"Больше информации \"%s\"",
					tasks[id].Title,
				),
				color.Black,
			),
			layout.NewSpacer(),
			widget.NewButtonWithIcon("",
				back,
				func() {
					w.SetContent(tasksContent)
					tasksList.Unselect(id)
				},
			),
		)

		taskTitle := widget.NewLabel(tasks[id].Title)
		taskTitle.TextStyle = fyne.TextStyle{Bold: true}

		taskDescription := widget.NewLabel(tasks[id].Description)
		taskDescription.TextStyle = fyne.TextStyle{Italic: true}

		taskDescription.Wrapping = fyne.TextWrapBreak

		buttonsBox := container.NewHBox(
			// Удаление задачи
			widget.NewButtonWithIcon(
				"",
				delete,
				func() {
					dialog.ShowConfirm(
						"Удалить запись ?",
						fmt.Sprintf("Вы действилтельно хотите удалить задачу  \"%s\" ?", tasks[id].Title),
						func(b bool) {
							if b {
								db.Delete(&Task{}, "id =?", tasks[id].Id)
								db.Find(&tasks)

								if len(tasks) == 0 {
									noTasksLabel.Show()
								} else {
									noTasksLabel.Hide()
								}
							}
							tasksList.UnselectAll()
							w.SetContent(tasksContent)
						},
						w,
					)
				},
			),

			// Редактирование задачи
			widget.NewButtonWithIcon(
				"",
				edit,
				func() {
					editBar := container.NewHBox(
						canvas.NewText(
							fmt.Sprintf(
								"Редактировать \"%s\"",
								tasks[id].Title),
							color.Black,
						),
						layout.NewSpacer(),

						widget.NewButtonWithIcon("", back, func() {
							w.SetContent(tasksContent)
							tasksList.Unselect(id)
						}),
					)

					editTitle := widget.NewEntry()
					editTitle.SetText(tasks[id].Title)

					editDescription := widget.NewMultiLineEntry()
					editDescription.SetText(tasks[id].Description)

					// Сохранение изменений
					editButton := widget.NewButtonWithIcon(
						"Сохранить",
						save,
						// Изменение записи в БД
						func() {
							db.Model(&Task{}).Where("id=?", tasks[id].Id).Updates(
								Task{
									Title:       editTitle.Text,
									Description: editDescription.Text,
								},
							)
							db.Find(&tasks)

							w.SetContent(tasksContent)
							tasksList.UnselectAll()
						},
					)

					editContent = container.NewVBox(
						editBar,
						canvas.NewLine(color.Black),

						editTitle,
						editDescription,
						editButton,
					)
					w.SetContent(editContent)
				},
			),
		)

		detailsVBox := container.NewVBox(
			detailsBar,
			canvas.NewLine(color.Black),

			taskTitle,
			taskDescription,
			buttonsBox,
		)
		w.SetContent(detailsVBox)
	}

	taskScroll := container.NewScroll(tasksList)
	taskScroll.SetMinSize(fyne.NewSize(500, 400))

	tasksContent = container.NewVBox(
		taskBar,
		canvas.NewLine(color.Black),
		noTasksLabel,
		taskScroll,
	)

	titleEntry := widget.NewEntry()
	titleEntry.SetPlaceHolder("Заголовок...")

	descriptionEntry := widget.NewMultiLineEntry()
	descriptionEntry.SetPlaceHolder("Подробное описание...")

	saveTaskButton := widget.NewButtonWithIcon(
		"Сохранить",
		save,
		func() {
			newTask := Task{Title: titleEntry.Text, Description: descriptionEntry.Text}
			db.Create(&newTask)
			db.Find(&tasks)

			titleEntry.Text = "" //очистка полей
			titleEntry.Refresh()

			descriptionEntry.Text = "" //очистка полей
			descriptionEntry.Refresh()

			w.SetContent(tasksContent)

			tasksList.UnselectAll()

			if len(tasks) == 0 {
				noTasksLabel.Show()
			} else {
				noTasksLabel.Hide()
			}
		})

	createBar := container.NewHBox(
		canvas.NewText("Новая задача", color.Black),
		layout.NewSpacer(),
		widget.NewButtonWithIcon("", back, func() {
			titleEntry.Text = "" //очистка полей
			titleEntry.Refresh()

			descriptionEntry.Text = "" //очистка полей
			descriptionEntry.Refresh()

			w.SetContent(tasksContent)

			tasksList.UnselectAll()
		}),
	)

	createContent = container.NewVBox(
		createBar,
		canvas.NewLine(color.Black),

		container.NewVBox(
			titleEntry,
			descriptionEntry,
			saveTaskButton,
		),
	)

	w.SetContent(tasksContent)

	w.Show()
	a.Run()
}
