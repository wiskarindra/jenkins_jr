class CreateCategories < ActiveRecord::Migration[5.1]
  def up
    create_table :categories do |t|
      t.integer :bukalapak_category_id, default: 0
      t.string  :bukalapak_category_name
      t.integer :count, default: 0

      t.timestamps null: false
    end

    add_index :categories, [:bukalapak_category_id], name: 'index_categories_on_bukalapak_category_id', unique: true
  end

  def down
    drop_table :categories
  end
end
