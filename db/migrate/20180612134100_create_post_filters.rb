class CreatePostFilters < ActiveRecord::Migration[5.1]
  def up
    create_table :post_filters do |t|
      t.integer :post_id
      t.integer :bukalapak_category_id, default: 0
    end

    add_index :post_filters, [:post_id, :bukalapak_category_id], name: "index_post_filters_on_post_id_and_bukalapak_category_id", unique: true
  end

  def down
    drop_table :post_filters
  end
end
