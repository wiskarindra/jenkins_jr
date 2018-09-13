class AddBukalapakCategoryIdToPostTags < ActiveRecord::Migration[5.1]
  def change
    add_column :post_tags, :bukalapak_category_id, :integer, default: 0
  end
end
